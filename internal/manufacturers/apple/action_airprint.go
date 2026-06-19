package apple

import (
	"encoding/binary"
	"encoding/hex"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// AirPrint decodes the AirPrint advertisement (subtype 0x03). Some fields
// remain underspecified upstream; we surface the raw bytes so the reporter
// can investigate later.
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/airprint.md
type AirPrint struct {
	AddressType   int    `json:"address_type"`
	PathType      int    `json:"path_type"`
	SecurityType  int    `json:"security_type"`
	Port          int    `json:"port"`
	Address       string `json:"address"` // 16 bytes; IPv6-shaped, hex-encoded
	MeasuredPower int    `json:"measured_power,omitempty"`
}

func processAirPrint(data []byte) []byte {
	type Apple struct {
		AirPrint `json:"airprint"`
	}

	if len(data) < 21 {
		return mf.MarshalOrEmpty(Apple{AirPrint{}})
	}

	ap := AirPrint{
		AddressType:  int(data[0]),
		PathType:     int(data[1]),
		SecurityType: int(data[2]),
		Port:         int(binary.BigEndian.Uint16(data[3:5])),
		Address:      hex.EncodeToString(data[5:21]),
	}
	if len(data) >= 22 {
		ap.MeasuredPower = int(int8(data[21]))
	}
	return mf.MarshalOrEmpty(Apple{ap})
}
