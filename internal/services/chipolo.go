package services

import (
	"encoding/hex"
)

// Chipolo Bluetooth tracker. Two service UUIDs observed in the wild:
//
//   - 0xfe65 — the primary Chipolo identifier (advertised UUID).
//   - 0xfe33 — secondary service-data block carrying a 10-byte payload:
//             4-byte header, followed by the device's own BD_ADDR. The
//             BD_ADDR echo is the headline privacy issue: a stable
//             identifier broadcast openly regardless of MAC randomization.
//
// The newer "Chipolo One Spot" / "Chipolo CARD Spot" also participate in
// Apple's Find My Network and broadcast Apple subtype 0x12 alongside;
// those are parsed by the Apple findmy handler.
type Chipolo struct {
	UUID     string `json:"uuid"`
	Header   string `json:"header,omitempty"`     // first 4 bytes hex (when 0xfe33)
	EchoMAC  string `json:"echo_mac,omitempty"`   // BD_ADDR echoed in the payload
	Raw      string `json:"raw,omitempty"`        // hex when format isn't recognised
}

func parseChipolo(data []byte, uuid string) []byte {
	c := Chipolo{UUID: uuid}
	switch {
	case len(data) == 10 && uuid == "fe33":
		c.Header = hex.EncodeToString(data[0:4])
		// BD_ADDR bytes appear in BLE wire order (LSB first); flip them
		// to canonical aa:bb:cc:dd:ee:ff for readability.
		c.EchoMAC = bleAddrString(data[4:10])
	default:
		c.Raw = hex.EncodeToString(data)
	}
	return raw(c)
}

// bleAddrString formats a 6-byte BLE address read from a wire-order buffer
// (little-endian — least significant byte first) into the canonical
// colon-separated hex form.
func bleAddrString(b []byte) string {
	if len(b) != 6 {
		return ""
	}
	const hexDigits = "0123456789abcdef"
	out := make([]byte, 0, 17)
	for i := 5; i >= 0; i-- {
		out = append(out, hexDigits[b[i]>>4], hexDigits[b[i]&0x0f])
		if i > 0 {
			out = append(out, ':')
		}
	}
	return string(out)
}
