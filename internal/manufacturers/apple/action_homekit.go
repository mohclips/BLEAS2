package apple

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// HomeKit decodes the HomeKit Accessory advertisement (subtype 0x06).
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/homekit.md
type HomeKit struct {
	StatusFlags         int    `json:"status_flags"`
	DeviceID            string `json:"device_id"`
	Category            int    `json:"category"`
	CategoryName        string `json:"category_name"`
	GlobalStateNumber   int    `json:"global_state_number"`
	ConfigurationNumber int    `json:"configuration_number"`
	CompatibleVersion   int    `json:"compatible_version"`
}

func processHomeKit(data []byte) []byte {
	type Apple struct {
		HomeKit `json:"homekit"`
	}

	if len(data) < 13 {
		return mf.MarshalOrEmpty(Apple{HomeKit{}})
	}

	hk := HomeKit{
		StatusFlags:         int(data[0]),
		DeviceID:            hex.EncodeToString(data[1:7]),
		Category:            int(binary.LittleEndian.Uint16(data[7:9])),
		GlobalStateNumber:   int(binary.LittleEndian.Uint16(data[9:11])),
		ConfigurationNumber: int(data[11]),
		CompatibleVersion:   int(data[12]),
	}
	if name, ok := homeKitCategories[hk.Category]; ok {
		hk.CategoryName = name
	} else {
		hk.CategoryName = fmt.Sprintf("unknown 0x%04x", hk.Category)
	}
	return mf.MarshalOrEmpty(Apple{hk})
}

// HomeKit Accessory Categories per Apple's HAP specification.
var homeKitCategories = map[int]string{
	1:  "Other",
	2:  "Bridge",
	3:  "Fan",
	4:  "Garage Door Opener",
	5:  "Lightbulb",
	6:  "Door Lock",
	7:  "Outlet",
	8:  "Switch",
	9:  "Thermostat",
	10: "Sensor",
	11: "Security System",
	12: "Door",
	13: "Window",
	14: "Window Covering",
	15: "Programmable Switch",
	16: "Range Extender",
	17: "IP Camera",
	18: "Video Doorbell",
	19: "Air Purifier",
	20: "Heater",
	21: "Air Conditioner",
	22: "Humidifier",
	23: "Dehumidifier",
	24: "Apple TV",
	25: "Speaker",
	26: "AirPort",
	27: "Sprinkler",
	28: "Faucet",
	29: "Shower Head",
	30: "Television",
	31: "Target Controller",
	32: "WiFi Router",
}
