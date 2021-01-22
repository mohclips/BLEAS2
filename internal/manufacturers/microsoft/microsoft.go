package microsoft

import (
	"encoding/hex"
	"encoding/json"

	//"log"
	log "github.com/mohclips/BLEAS2/internal/logging"

	"github.com/mohclips/BLEAS2/internal/utils"
)

// see https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-cdp/77b446d0-8cea-4821-ad21-fabdf4d9a569

// Beacon Data (24 bytes): The beacon data section is further broken down.
// Note that the Scenario and Subtype Specific Data section requirements will differ based on the Scenario and Subtype.

// byte 0 = Scenario Type = Scenario Type (1 byte): Set to 1
// byte 1 = Version and Device Type (see below) -  The high two bits are set to 00 for the version number; the lower6 bits are set to Device Type
// byte 2 = Version and Flags = Version and Flags (1 byte): The high 3 bits are set to 001; the lower 3 bits to 00000.
// byte 3 = Reserved = Reserved (1 byte): Currently set to zero.
// bytes 4-7 = Salt = Salt (4 bytes): Four random bytes.
// bytes Device Hash (24 bytes) = Device Hash (24 bytes): SHA256 Hash of Salt plus Device Thumbprint. Truncated to 16 bytes.

var microsoftDevice = map[int]string{
	1:  "Xbox One",
	6:  "Apple iPhone",
	7:  "Apple iPad",
	8:  "Android device",
	9:  "Windows 10 Desktop",
	11: "Windows 10 Phone",
	12: "Linus device",
	13: "Windows IoT",
	14: "Surface Hub",
}

// https://docs.microsoft.com/en-us/uwp/api/windows.devices.bluetooth.advertisement.bluetoothleadvertisementflags?view=winrt-19041
var bleFlags = map[int]string{
	16: "DualModeHostCapable",       // Specifies simultaneous Bluetooth LE and BR/EDR to same device capable (host)
	8:  "DualModeControllerCapable", //Specifies simultaneous Bluetooth LE and BR/EDR to same device capable (controller).
	4:  "ClassicNotSupported",       //Specifies Bluetooth BR/EDR not supported.
	2:  "GeneralDiscoverableMode",   //Specifies Bluetooth LE General Discoverable Mode.
	1:  "LimitedDiscoverableMode",   //Specifies Bluetooth LE Limited Discoverable Mode.
	0:  "None",                      //Specifies no flag.
}

// ParseMF - parse manufacturers data
func ParseMF(mmfData []byte) string {

	// mfData contains the Maufacturer ID two bytes at start
	mfData := mmfData[2:]

	scenarioType := mfData[0] // Set to 1

	deviceByte := mfData[1] // The high two bits are set to 00 for the version number; the lower6 bits are set to Device Type
	version := (deviceByte & 0b11000000) >> 6
	deviceType := deviceByte & 0b00111111

	var deviceName string
	var found bool
	if deviceName, found = microsoftDevice[int(deviceType)]; !found {
		deviceName = "Unknown"
	}

	versionFlags := mfData[2]              // The high 3 bits are set to 001; the lower 3 bits to 00000.
	vfVersion := versionFlags & 0b11100000 // always 32 ?
	vfFlags := (versionFlags & 0b00011111) >> 5
	flags := utils.BitmaskToNames(int(vfFlags), bleFlags)

	reserved := mfData[3] // should be zero - but isnt

	rsalt := mfData[4:8] // Salt (4 bytes): Four random bytes.
	salt := utils.FormatHexComma(hex.EncodeToString(rsalt))

	rsha256 := mfData[8:] // Device Hash (24 bytes): SHA256 Hash of Salt plus Device Thumbprint. Truncated to 16 bytes.
	sha256 := utils.FormatHexComma(hex.EncodeToString(rsha256))

	// Microsoft - packet
	type Microsoft struct {
		ScenarioType  int      `json:"scenario_type"`  // Note that Marshall only exports "exportable" names, that is not lowercase
		DeviceVersion int      `json:"device_version"` // underscore is not a vaild property name, can only be used in json metadata
		DeviceType    int      `json:"device_type"`
		DeviceName    string   `json:"device_name"`
		VersionFlags  int      `json:"version_flags"`
		Version       int      `json:"version"`
		RawFlags      int      `json:"raw_flags"`
		Flags         []string `json:"flags"`
		Reserved      int      `json:"reserved"`
		Salt          string   `json:"salt"`
		Sha256        string   `json:"sha256"`
	}

	pkt := Microsoft{
		ScenarioType:  int(scenarioType),
		DeviceVersion: int(version), // should be always zero - BLE version?
		DeviceType:    int(deviceType),
		DeviceName:    deviceName,
		VersionFlags:  int(versionFlags),
		Version:       int(vfVersion), // should always be 32, need to check
		RawFlags:      int(vfFlags),
		Flags:         flags,
		Reserved:      int(reserved), // supposedly set to zero, but its not!
		//RawReserved:     int(reserved),     // supposedly set to zero, but its not!
		Salt:   salt,
		Sha256: sha256,
	}

	// convert to json
	var mpkt []byte
	var err error
	mpkt, err = json.Marshal(pkt)
	if err != nil {
		log.Error("%s", err)
		mpkt = nil
	}

	ret := "{\"microsoft\":" + string(mpkt) + "}"

	return ret
}
