# CUB Tire Insight-BLE App II — Reverse Engineering Report

## Target

| Field | Value |
|---|---|
| App Name | TIRE INSIGHT-BLE APP II |
| Package | `cub.tireinsightService` |
| Version | 3.1.5 (vercode 20) |
| Size | 29 MB |
| MD5 | `9d75702806b2eb15bebf4186ab2efb28` |
| Developer | CUB Elecparts Inc. (`com.cubelec`) |
| Platform | Android 7.1+ (Nougat) |
| Build Source | `C:/Users/a1937/Desktop/work/CUB project/tireInsightBLE_rework/` |
| APK Source | `https://pool.apk.aptoide.com/aptoide-web/cub-tireinsightservice-20-73714697-9d75702806b2eb15bebf4186ab2efb28.apk` |

## Analysis Artifacts

| Artifact | Path |
|---|---|
| Original APK | `mobile/apks/cub.tireinsightService_v3.1.5.apk` |
| apktool (smali + resources) | `mobile/targets/tire-insight-ble-ii/apktool/` |
| jadx (decompiled Java) | `mobile/targets/tire-insight-ble-ii/jadx/` |

---

## App Architecture

The app is a **Flutter application** compiled to `libapp.so` (arm64-v8a / armeabi-v7a / x86_64). The Dart layer handles all BLE logic and passes down to Android via Flutter MethodChannels. Two parallel BLE protocols are used simultaneously.

```
Flutter (Dart / libapp.so)
    ├── dchs_flutter_beacon  ──→  AltBeacon (org.altbeacon)   ──→ iBeacon ranging
    └── flutter_blue_plus    ──→  FlutterBluePlusAndroid (V0/j.java) ──→ GATT connect
```

---

## BLE Architecture — Two-Layer Approach

| Layer | Library | Purpose |
|---|---|---|
| iBeacon Ranging | `dchs_flutter_beacon` (AltBeacon) | Passive BLE advertisement scan — **main TPMS data channel** |
| GATT Connect | `flutter_blue_plus` | Direct device connection (configuration, OTA, etc.) |

---

## iBeacon Proximity UUID (Primary Sensor Identifier)

```
b54adc00-67f9-11d9-9669-0800200c9a66
```

This is the **iBeacon proximity UUID** broadcast by all CUB Elecparts TPMS sensors. The app ranges against this UUID via `startRangingBeaconsInRegion()`. Every sensor beacon in range is returned as a ranging result — no pairing or GATT connection required for live tire data.

---

## Sensor Data Connection Flow

```
startRangingBeaconsInRegion(
    proximityUUID: b54adc00-67f9-11d9-9669-0800200c9a66
)
    |
    v
iBeacon advertisement received
    { proximityUUID, major, minor, rssi, accuracy, macAddress }
    |
    v
BeaconDataProcessor._parseBeaconData()
    |
    ├── _parseMajorMinor(major, minor)
    |       ├── pressureFromMinor(minor)      ← pressure encoded in minor word
    |       ├── temperatureFromMinor(minor)   ← temp extracted via _minorFirstBit mask
    |       └── _toBitMask(minor)             ← alert/status flags from minor bits
    |
    v
TireData {
    pressureKpa,
    tempC,
    alertType  { lowPressure, highPressure, highTemperature, lowBattery },
    rssi,
    sensorId   ← derived from major
}
    |
    v
BeaconReading → TireDataModel → UI
```

---

## Sensor Data Encoding (iBeacon Major / Minor) — CONFIRMED FROM DISASSEMBLY

All tire data is carried in the **16-bit iBeacon `minor`** word. The `major` word
identifies the sensor. The decode was recovered by disassembling the six
`BeaconBindingUtils` methods in `libapp.so` (Dart AOT, arm64). The shared helper
`_to4Hex(minor)` renders the 16-bit minor as a 4-hex-digit string `"HHHH"`
(`minor.toRadixString(16).padLeft(4,'0')`); the temperature and pressure routines
then slice that string into the **high byte** (chars 0–1) and **low byte**
(chars 2–3) and parse each back to an int.

```
minor (16 bits) = 0x AB CD
                     │  │  └── low  byte (bits 7..0)   → PRESSURE
                     │  └───── high byte (bits 15..8)  → TEMPERATURE (+ flag in MSB)
                     └──────── bit 15 = status/alarm flag  (_minorFirstBit)
```

### Decode formulas (verbatim from libapp.so)

| Quantity | Source bits | Formula | Code |
|---|---|---|---|
| **Pressure** | low byte `minor & 0xFF` | `pressure_kPa = (minor & 0xFF) * 2.5` | `pressureFromMinor` @ `0x32364c` (`SCVTF`/`FMOV D1,#2.5`/`FMUL`) |
| **Temperature** | high byte, low 7 bits | `temp_C = ((minor >> 8) & 0x7F) - 40`  (if result == 86 → clamp to 85) | `temperatureFromMinor` @ `0x323798` (`AND W1,W0,#0x7f`/`SUB X2,X1,#0x28`/`CMP #0x56`) |
| **Status flag** | bit 15 (MSB of high byte) | `flag = (minor >> 15) & 1` | `_minorFirstBit` @ `0x579b18` |

Worked example: `minor = 0x4B5A`
- Pressure = `0x5A` × 2.5 = 90 × 2.5 = **225 kPa** (≈ 32.6 psi / 2.25 bar)
- Temperature = `(0x4B & 0x7F)` − 40 = 75 − 40 = **35 °C**
- Status bit = `0x4B >> 7` = 0 → no alarm

Unit conversions: `psi = kPa × 0.145038`, `bar = kPa / 100`.
Pressure range: 0–255 → 0–637.5 kPa (0–92 psi). Temperature range: −40 to +87 °C (clamped at 85).

### Call chain (confirmed)

```
BeaconDataProcessor._parseBeaconData   @ 0x322e80
    LDUR X4,[X0,#23]                    ; load Beacon.minor (dchs Beacon, tagged offset +23)
    BL   parseTempPressure              ; 0x322ebc + 0x710 → 0x3235cc
BeaconBindingUtils.parseTempPressure   @ 0x3235cc
    BL   temperatureFromMinor(minor)    ; → 0x323798, result saved
    BL   pressureFromMinor(minor)       ; → 0x32364c
    return (pressure, temperature)      ; Smi-tagged pair: X0=pressure, X1=temperature
```

### Other `BeaconBindingUtils` helpers (recovered names)

| Method | Address | Purpose |
|---|---|---|
| `_to4Hex` | `0x323704` | 16-bit minor → 4-digit hex string (substring source for the two decoders) |
| `_minorFirstBit` | `0x579b18` | MSB of minor = status/alarm flag |
| `_macLastByte` | `0x32392c` | last octet of sensor MAC (parsed from hex string) |
| `generateBindingId` | `0x32381c` | builds sensor binding ID from MAC + minor hex |

> Note: the dchs_flutter_beacon `Beacon` object stores `minor` at tagged offset +23
> (8-byte header + 4-byte compressed pointers). The app's own serialization model
> (`Beacon.fromJson` @ `0x36fb08`) uses a different layout
> (`macAddress`+7, `major`+11, `minor`+15, `rssi`+23) — don't confuse the two.

---

## Wheel Position Mapping

Sensors are bound to vehicle positions via `Beacon.applyVehicleBindings()` and `_updateSensorMapping()`:

| Field | Position |
|---|---|
| `frontId` | Front (3-wheel / single front) |
| `frontRightId` | Front Right |
| `rearLeftId` | Rear Left |
| `rearRightId` | Rear Right |
| `middleMaxPsi` / `middleMinPsi` | Middle axle (6-wheel vehicles) |

Supported vehicle types: `four_wheel`, `six_wheel`, `three_wheel` (tricycle).

Binding trace logging:
```
[VEH-TRACE] Beacon._updateSensorMapping frontId=...
[VEH-TRACE] Beacon.applyVehicleBindings(ENTER) frontId=...
[VEH-TRACE] Beacon.applyVehicleBindings(AFTER_NORMALIZE) frontId=...
```

---

## GATT Connection (FlutterBluePlus)

The GATT layer (`V0/j.java` — `FlutterBluePlusAndroid`) handles direct device connections. No custom proprietary GATT service UUIDs are hardcoded — sensors operate primarily as iBeacon advertisers.

### Standard BLE UUIDs observed

| Service UUID | Characteristic UUID | Purpose |
|---|---|---|
| `1801` | `2A05` | GATT Service Changed (standard) |

### UUID format helper (`V0/j.java:uuidStr()`)

Short UUIDs (4 chars) are expanded to full 128-bit using the Bluetooth base UUID:
```
0000XXXX-0000-1000-8000-00805f9b34fb
```

### GATT Callback events (V0/i.java)

| Event | Handler |
|---|---|
| `onCharacteristicChanged` | `OnCharacteristicReceived` |
| `onCharacteristicRead` | `OnCharacteristicReceived` |
| `onCharacteristicWrite` | logged, forwarded to Dart |
| `onDescriptorRead` | `OnDescriptorRead` |
| `onReadRssi` | `OnReadRssi` |

---

## Android BLE Scan Settings

Configured in `cub.tireinsightService.BleScanManager` (`BleScanManager.java`):

| Setting | Value | Meaning |
|---|---|---|
| Scan mode | `LOW_LATENCY` (2) | ~5ms scan interval — highest frequency |
| Callback type | `FIRST_MATCH` (1) | Report each device once on first detection |
| Match mode | `AGGRESSIVE` (1) | Match with fewer packets |
| Report delay | `0ms` | Immediate delivery, no batching |

Android 15+ (API 35+) specific high-frequency scan path is handled separately with explicit `ScanSettings` tuning.

---

## Flutter Method Channels

| Channel Name | Direction | Purpose |
|---|---|---|
| `cub.tireinsightService/ble_scan` | Dart → Java | `configureHighFrequencyScan`, `checkBluetoothStatus` |
| `cub.tireinsightService/screen_state` | Java → Dart | Screen on/off events for background scanning control |
| `com.cubelec.bletpms2/vibration` | Dart → Java | Alert vibration: `vibrate`, `startWaveform`, `cancel` |
| `com.cubelec.bletpms2/exact_alarm` | Dart → Java | `isExactAlarmAllowed`, `requestExactAlarm` (Android 12+) |
| `tpms_foreground_channel` | Dart ↔ Java | Foreground service lifecycle |
| `flutter_beacon` | Dart ↔ Java | iBeacon ranging control |
| `flutter_beacon_event` | Java → Dart | iBeacon ranging results stream |

---

## Key Dart Classes (libapp.so symbols)

| Symbol | Role |
|---|---|
| `BeaconDataProcessor._internal@720062628` | Singleton processor for incoming beacon data |
| `BeaconDataProcessor._parseBeaconData@720062628` | Decodes raw iBeacon major/minor into TireData |
| `_parseMajorMinor@547358469` | Extracts pressure, temp, status from major/minor |
| `pressureFromMinor` | Converts minor value → pressure in kPa |
| `temperatureFromMinor` | Converts minor value → temperature in °C |
| `_minorFirstBit@699045692` | Bitmask helper for minor field parsing |
| `_toBitMask@323082469` | Converts integer to alert flag bitmask |
| `TireData` / `TireDataModel` / `TireDataBridge` | Data model layers |
| `Beacon.applyVehicleBindings` | Maps sensor IDs to wheel positions |
| `Beacon._updateSensorMapping` | Updates front/rear/left/right sensor ID table |
| `_restartAndroidRangingForTargetUuid@710333139` | Restarts ranging on connection drop |
| `_startScanHealthCheck@703205134` | Watchdog to detect and recover stalled scans |
| `smartRestartScanning` | Graceful scan restart logic |
| `processScanResult` | Handles incoming FBP BLE scan results |

---

## App Routes (Navigation)

```
/app_flow
/pressure_relief_learning
/pressure_relief_learning_four_wheeled
/pressure_relief_learning_six_wheeled
/pressure_relief_learning_tricycle
/add_vehicle_pressure_relief_learning
/add_vehicle_pressure_relief_learning_four_wheeled
/add_vehicle_pressure_relief_learning_six_wheeled
/add_vehicle_pressure_relief_learning_tricycle
```

---

## Localization

Supported languages (from `assets/translations/`):

`en`, `zh-TW`, `zh-CN`, `ja`, `de`, `es`

---

## Summary

The CUB Tire Insight-BLE App II connects to TPMS sensors **exclusively via iBeacon passive ranging** using the proximity UUID `b54adc00-67f9-11d9-9669-0800200c9a66`. Sensors do not require pairing or GATT connections — all pressure, temperature, and alert data is encoded in the **iBeacon major/minor fields** of the BLE advertisement. The minor word carries pressure (kPa), temperature (°C), and status flags via bitmask. The major value identifies the individual sensor. The app maps sensor IDs to wheel positions (FL/FR/RL/RR + middle for 6-wheel) and alerts the driver via audio/vibration when thresholds are exceeded.
