-- CUB Elecparts TPMS decoder (Nick's van + any other CUB-family vehicle
-- that drives past).
--
-- Background — see CUB-TIRE-INSIGHT-REPORT.md:
--   * All CUB Elecparts TPMS sensors broadcast as iBeacons with the
--     fixed proximity UUID below. No GATT, no service data — just an
--     iBeacon advert with sensor data packed into major/minor.
--   * `major` (16-bit) = lower 16 bits of the 28-bit sensor ID printed
--     in the Tire Insight app. Sufficient to disambiguate 4 wheels on
--     one vehicle.
--   * `minor` (16-bit) = pressure (kPa) + temperature (°C) + alert flags,
--     bit-packed. Exact layout needs libapp.so disassembly of
--     pressureFromMinor() / temperatureFromMinor() — TODO.
--   * Sensors only advertise sparsely (centrifugal-switch wake when wheel
--     spins, plus state-change events). Don't expect a steady stream.
--
-- Nick's Nissan van sensor ID lower-16-bits ↔ wheel mapping:
--   LF aa0a31f → 0xa31f = 41759  (Left Front)
--   RF 510a4fb → 0xa4fb = 42235  (Right Front)
--   LR 1c0a0e6 → 0xa0e6 = 41190  (Left Rear)
--   RR a509a0c → 0x9a0c = 39436  (Right Rear)  ← already confirmed

SET VARIABLE cub_uuid = 'b54adc00-67f9-11d9-9669-0800200c9a66';

-- Materialize CUB broadcasts + wheel mapping into a temp view so we can
-- emit both a per-wheel summary and the broadcast timeline against it.
-- CTEs are scoped to a single statement in DuckDB, so a view is needed
-- when we want to issue multiple SELECTs over the same derived data.
CREATE OR REPLACE TEMP VIEW cub_with_wheel AS
WITH raw AS (
    SELECT
        oj.first_seen                                          AS ts,
        oj.address                                             AS mac,
        oj.address_type                                        AS addr_type,
        o.rssi_mean::INT                                       AS rssi,
        o.sample_count                                         AS samples,
        json_extract(oj.mfr_details_json,
                     '$.apple.ibeacon.major')::BIGINT          AS major_,
        json_extract(oj.mfr_details_json,
                     '$.apple.ibeacon.minor')::BIGINT          AS minor_
    FROM obs o
    JOIN obs_json oj USING (address, first_seen)
    WHERE json_extract_string(oj.mfr_details_json,
                              '$.apple.ibeacon.uuid')
          = getvariable('cub_uuid')
)
-- Decode pressure / temperature / status flag from the 16-bit minor.
-- Formulas confirmed from libapp.so disassembly (see
-- CUB-TIRE-INSIGHT-REPORT.md §Sensor Data Encoding):
--   minor = 0xAB CD  where AB = temp byte, CD = pressure byte
--   pressure_kPa = (minor & 0xFF) * 2.5
--   temp_C       = ((minor >> 8) & 0x7F) - 40    (clamp 86 → 85)
--   status_flag  = (minor >> 15) & 1
SELECT
    ts, mac, addr_type, rssi, samples, major_, minor_,
    CASE major_
        WHEN 41759 THEN 'Van LF (Left Front)'
        WHEN 42235 THEN 'Van RF (Right Front)'
        WHEN 41190 THEN 'Van LR (Left Rear)'
        WHEN 39436 THEN 'Van RR (Right Rear)'
        ELSE            '(unknown vehicle/sensor)'
    END                                                    AS wheel,
    ((minor_ & 255) * 2.5)::DECIMAL(6,1)                   AS pressure_kpa,
    round((minor_ & 255) * 2.5 * 0.01, 2)::DECIMAL(5,2)    AS pressure_bar,
    round((minor_ & 255) * 2.5 * 0.145038, 1)::DECIMAL(5,1) AS pressure_psi,
    CASE
        WHEN ((minor_ >> 8) & 127) - 40 = 86 THEN 85
        ELSE                                      ((minor_ >> 8) & 127) - 40
    END::INT                                               AS temp_c,
    ((minor_ >> 15) & 1)::INT                              AS status_flag
FROM raw;

-- Block 1 — per-wheel summary. All four expected van sensors, left-joined
-- against what we caught, so a missing wheel shows up explicitly as ❌.
SELECT 'WHEEL SUMMARY' AS block;
WITH expected_wheels(major_, wheel) AS (
    VALUES
        (41759, 'Van LF (Left Front)'),
        (42235, 'Van RF (Right Front)'),
        (41190, 'Van LR (Left Rear)'),
        (39436, 'Van RR (Right Rear)')
)
SELECT
    CASE WHEN count(c.ts) > 0 THEN '✅ seen'
         ELSE                       '❌ MISSING'
    END                                                   AS status,
    ew.wheel,
    printf('0x%04x', ew.major_)                           AS major_hex,
    coalesce(any_value(c.mac),
             '(unknown — sensor never captured)')         AS mac,
    count(c.ts)                                           AS broadcasts,
    coalesce(sum(c.samples), 0)                           AS samples,
    coalesce(max(c.rssi)::VARCHAR, '')                    AS rssi_best,
    coalesce(min(c.rssi)::VARCHAR, '')                    AS rssi_worst,
    min(c.ts)                                             AS first_seen,
    max(c.ts)                                             AS last_seen
FROM expected_wheels ew
LEFT JOIN cub_with_wheel c ON c.major_ = ew.major_
GROUP BY ew.major_, ew.wheel
ORDER BY status DESC, ew.wheel;

-- Block 2 — full timeline (every broadcast, oldest → newest).
SELECT 'BROADCAST TIMELINE' AS block;
SELECT
    ts,
    wheel,
    pressure_kpa                                      AS kpa,
    pressure_bar                                      AS bar,
    pressure_psi                                      AS psi,
    temp_c                                            AS "temp_°C",
    CASE WHEN status_flag = 1 THEN 'ALARM' ELSE 'ok' END AS status,
    rssi,
    samples,
    printf('0x%04x', minor_)                          AS minor_hex,
    CASE
        WHEN rssi > -65 THEN 'office (scanner room)'
        WHEN rssi > -80 THEN 'directly downstairs'
        WHEN rssi > -94 THEN 'downstairs far / driveway'
        ELSE                 'outside / next door / passing'
    END                                               AS location
FROM cub_with_wheel
ORDER BY ts;
