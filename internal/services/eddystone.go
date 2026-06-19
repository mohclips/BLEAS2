package services

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// Eddystone is Google's open BLE beacon format (service UUID 0xfeaa). Four
// frame types share the same service-data prefix:
//
//   - 0x00 UID  — 10-byte namespace + 6-byte instance
//   - 0x10 URL  — compressed URL prefix + suffix bytes
//   - 0x20 TLM  — telemetry: battery voltage, temperature, advert count, uptime
//   - 0x30 EID  — encrypted ephemeral identifier
//
// See: https://github.com/google/eddystone

type Eddystone struct {
	FrameType    string `json:"frame_type"`              // "uid" | "url" | "tlm" | "eid" | "unknown"
	TXPower      int    `json:"tx_power,omitempty"`      // calibrated TX power at 0m, dBm (signed)
	Namespace    string `json:"namespace,omitempty"`     // UID frame: 10-byte namespace hex
	Instance     string `json:"instance,omitempty"`      // UID frame: 6-byte instance hex
	URL          string `json:"url,omitempty"`           // URL frame: decoded URL
	BatteryMV    int    `json:"battery_mv,omitempty"`    // TLM: mV
	Temperature  string `json:"temperature,omitempty"`   // TLM: signed 8.8 fixed-point
	AdvCount     uint32 `json:"adv_count,omitempty"`     // TLM: PDU count since boot
	UptimeS      uint32 `json:"uptime_s,omitempty"`      // TLM: 0.1s units → seconds
	EphemeralID  string `json:"ephemeral_id,omitempty"`  // EID frame: 8-byte hex
	Raw          string `json:"raw,omitempty"`           // hex when frame type unknown
}

func parseEddystone(data []byte) []byte {
	if len(data) < 2 {
		return raw(Eddystone{FrameType: "truncated"})
	}
	var e Eddystone
	switch data[0] {
	case 0x00: // UID
		if len(data) < 18 {
			e.FrameType = "uid_truncated"
			break
		}
		e.FrameType = "uid"
		e.TXPower = int(int8(data[1]))
		e.Namespace = hex.EncodeToString(data[2:12])
		e.Instance = hex.EncodeToString(data[12:18])
	case 0x10: // URL
		if len(data) < 3 {
			e.FrameType = "url_truncated"
			break
		}
		e.FrameType = "url"
		e.TXPower = int(int8(data[1]))
		e.URL = decodeEddystoneURL(data[2], data[3:])
	case 0x20: // TLM
		if len(data) < 14 {
			e.FrameType = "tlm_truncated"
			break
		}
		e.FrameType = "tlm"
		// data[1] = version; data[2:4]=battery mV, data[4:6]=temp,
		// data[6:10]=adv count, data[10:14]=uptime in 0.1s units
		e.BatteryMV = int(binary.BigEndian.Uint16(data[2:4]))
		e.Temperature = fmt.Sprintf("%d.%02d", int8(data[4]), data[5])
		e.AdvCount = binary.BigEndian.Uint32(data[6:10])
		e.UptimeS = binary.BigEndian.Uint32(data[10:14]) / 10
	case 0x30: // EID
		if len(data) < 10 {
			e.FrameType = "eid_truncated"
			break
		}
		e.FrameType = "eid"
		e.TXPower = int(int8(data[1]))
		e.EphemeralID = hex.EncodeToString(data[2:10])
	default:
		e.FrameType = fmt.Sprintf("unknown_0x%02x", data[0])
		e.Raw = hex.EncodeToString(data)
	}
	return raw(e)
}

// decodeEddystoneURL applies the Eddystone-URL prefix/suffix compression.
// https://github.com/google/eddystone/tree/master/eddystone-url
func decodeEddystoneURL(scheme byte, body []byte) string {
	prefixes := []string{"http://www.", "https://www.", "http://", "https://"}
	if int(scheme) >= len(prefixes) {
		return fmt.Sprintf("?scheme=0x%02x:%s", scheme, hex.EncodeToString(body))
	}
	suffixes := map[byte]string{
		0x00: ".com/", 0x01: ".org/", 0x02: ".edu/", 0x03: ".net/",
		0x04: ".info/", 0x05: ".biz/", 0x06: ".gov/",
		0x07: ".com", 0x08: ".org", 0x09: ".edu", 0x0a: ".net",
		0x0b: ".info", 0x0c: ".biz", 0x0d: ".gov",
	}
	var b []byte
	b = append(b, prefixes[scheme]...)
	for _, c := range body {
		if s, ok := suffixes[c]; ok {
			b = append(b, s...)
			continue
		}
		b = append(b, c)
	}
	return string(b)
}
