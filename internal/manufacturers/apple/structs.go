package apple

// UnknownPacket - save for later
type UnknownPacket struct {
	Action int    `json:"action"`
	Len    int    `json:"len"`
	Data   []byte `json:"data"` // this gets Base64 encoded by Marshall
}

// Nearby - packet
type Nearby struct {
	State    string   `json:"state"`  // Note that Marshall only exports "exportable" names, that is not lowercase
	RawState int      `json:"_state"` // underscore is not a vaild property name, can only be used in json metadata
	Masks    []string `json:"masks"`
	RawMask  int      `json:"_mask"`
	Wifi     string   `json:"wifi"`
	RawWifi  int      `json:"_wifi"`
	Os       string   `json:"os"`
}

// Handoff - packet
type Handoff struct {
	Clipboard bool   `json:"clipboard"`
	SeqNum    uint16 `json:"seq_num"`
	GcmAuth   int    `json:"gcm_auth"`
	Payload   []byte `json:"payload"`
}

type sBattery struct {
	R int `json:"batteryR"`
	L int `json:"batteryL"`
}

type sCharging struct {
	Case int `json:"C"`
	R    int `json:"R"`
	L    int `json:"L"`
}

// Airpods - action packet
type Airpods struct {
	RawUndef1      int       `json:"_undef1"`
	RawDeviceModel uint16    `json:"_device_model"`
	DeviceModel    string    `json:"device_model"`
	RawStatus      int       `json:"_status"`
	RawBatteryRL   int       `json:"_batteryRL"`
	Battery        sBattery  `json:"battery"`
	RawPower       int       `json:"_power"`
	Charging       sCharging `json:"charging"`
	CasePower      int       `json:"case_power"`
	Lid            int       `json:"_lid"`
	RawColor       int       `json:"_color"`
	RawUndef2      int       `json:"_undef2"`
	RawPayload     []byte    `json:"_payload"`
}

// https://github.com/furiousMAC/continuity

// https://github.com/hexway/apple_bleee/blob/1f8022959be660b561e6004b808dd93fa252bc90/ble_read_state.py#L387

//
var blePacketsTypes = map[int]string{
	0:    "none",
	0x01: "unknown 0x01",
	0x02: "unknown 0x02",
	0x03: "airprint", // https://github.com/furiousMAC/continuity/blob/master/messages/airprint.md
	0x04: "unknown 0x04",
	0x05: "airdrop",
	0x06: "homekit",        // https://github.com/furiousMAC/continuity/blob/master/messages/homekit.md
	0x07: "airpods",        // https://github.com/furiousMAC/continuity/blob/master/messages/proximity_pairing.md
	0x08: "hey_siri",       // https://github.com/furiousMAC/continuity/blob/master/messages/hey_siri.md
	0x09: "airplay",        // https://github.com/furiousMAC/continuity/blob/master/messages/airplay_target.md
	0x0a: "airplay_source", // https://github.com/furiousMAC/continuity/blob/master/messages/airplay_source.md
	0x0b: "watch_c",        // https://github.com/furiousMAC/continuity/blob/master/messages/magic_switch.md
	0x0c: "handoff",        // https://github.com/furiousMAC/continuity/blob/master/messages/handoff.md
	0x0d: "wifi_set",       // https://github.com/furiousMAC/continuity/blob/master/messages/tethering_target.md
	0x0e: "hotspot",        // https://github.com/furiousMAC/continuity/blob/master/messages/tethering_source.md
	0x0f: "nearby_action",  // https://github.com/furiousMAC/continuity/blob/master/messages/nearby_action.md
	0x10: "nearby",         // https://github.com/furiousMAC/continuity/blob/master/messages/nearby_info.md
}

// https://github.com/hexway/apple_bleee/blob/1f8022959be660b561e6004b808dd93fa252bc90/ble_read_state.py//L107
// Activity Level codes - https://github.com/furiousMAC/continuity/blob/master/messages/nearby_info.md

// phoneStates
var phoneStates = map[int]string{
	0x00: "Activity level is not known",
	0x01: "Activity reporting is disabled",
	0x02: "unknown 0x02",
	0x03: "User is idle",
	0x04: "unknown 0x04",
	0x05: "Audio is playing with the screen off",
	0x06: "unknown 0x06",
	0x07: "Screen is on",
	0x08: "unknown 0x08",
	0x09: "Screen on and video playing",
	0x0a: "Watch is on wrist and unlocked",
	0x0b: "Recent user interaction",
	0x0c: "unknown 0x0c",
	0x0d: "User is driving a vehicle",
	0x0e: "Phone call or Facetime",
	0x0f: "unknown 0x0f",
	//NOTE: removed and usurped by status flags
	// 0x11: "Home screen",
	// 0x13: "Off",
	// 0x17: "Lock screen",
	// 0x18: "Off",
	// 0x1a: "Off",
	// 0x1b: "Home screen",
	// 0x1c: "Home screen",
	// 0x23: "Off",
	// 0x47: "Lock screen",
	// 0x4b: "Home screen",
	// 0x4e: "Outgoing call",
	// 0x57: "Lock screen",
	// 0x5a: "Off",
	// 0x5b: "Home screen",
	// 0x5e: "Outgoing call",
	// 0x67: "Lock screen",
	// 0x6b: "Home screen",
	// 0x6e: "Incoming call",
}

var nearbyStatusMasks = map[int]string{
	0:    "none",
	0x01: "AirPods are connected and the screen is on",
	0x02: "Authentication Tag is 4 bytes",
	0x04: "WiFi is on",
	0x08: "Unknown",
	0x10: "Authentication Tag is present",
	0x20: "Apple Watch is locked or not",
	0x40: "Auto Unlock on the Apple Watch is enabled",
	0x80: "Auto Unlock is enabled",
}

var nearbyActionTypes = map[int]string{
	0:    "none",
	0x01: "Apple TV Setup",
	0x04: "Mobile Backup",
	0x05: "Watch Setup",
	0x06: "Apple TV Pair",
	0x07: "Internet Relay",
	0x08: "WiFi Password",
	0x09: "iOS Setup",
	0x0A: "Repair",
	0x0B: "Speaker Setupd",
	0x0C: "Apple Pay",
	0x0D: "Whole Home Audio Setup",
	0x0E: "Developer Tools Pairing Request",
	0x0F: "Answered Call",
	0x10: "Ended Call",
	0x11: "DD Ping",
	0x12: "DD Pong",
	0x13: "Remote Auto Fill",
	0x14: "Companion Link Proximity",
	0x15: "Remote Management",
	0x16: "Remote Auto Fill Pong",
	0x17: "Remote Display",
}

var heySiriDevices = map[uint16]string{
	0:      "none",
	0x0002: "iPhone",
	0x0003: "iPad",
	0x0009: "MacBook",
	0x000A: "Watch",
}

// https://support.apple.com/en-gb/HT209580

// AirPods Max
// Model number: A2096
// Year introduced: 2020

// AirPods Pro
// Model number: A2084, A2083
// Year introduced: 2019

// AirPods (2nd generation)
// Model number: A2032, A2031
// Year introduced: 2019

// AirPods (1st generation)
// Model number: A1523, A1722
// Year introduced: 2017

var airpodDevices = map[uint16]string{
	0:      "none",
	0x0002: "iPhone",
	0x0003: "iPad",
	0x0009: "MacBook",
	0x000A: "Watch",
	0x0e20: "AirPods Pro",
	0x0f20: "Unknown model 0x0f20", // known unknown :) - prob. AirPods Max
}
