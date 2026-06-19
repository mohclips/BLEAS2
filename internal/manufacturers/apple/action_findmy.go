package apple

import (
	"encoding/hex"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// FindMy decodes Apple Find My Network advertisements (subtype 0x12). Two
// variants are emitted by Apple devices and accessories:
//
//   - "separated": 25-byte payload broadcast by a device that has lost contact
//     with its owner. Carries the bulk of the rotating EC P-224 public key,
//     a hint byte, and a status byte.
//   - "nearby": 2-byte payload with just the status and public-key bits.
//     Emitted when the device is still in proximity to its paired owner.
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/findmy.md
// and the v4.4.0 Wireshark dissector for the 2-byte variant.
type FindMy struct {
	Variant       string `json:"variant"`                 // "nearby" | "separated" | "raw"
	Status        int    `json:"status"`                  // status byte
	Maintained    bool   `json:"maintained"`              // bit 2 set => owner connected
	BatteryLevel  string `json:"battery_level"`           // bits 6-7: full | medium | low | critical
	PublicKeyBits int    `json:"public_key_bits"`         // low 2 bits encode bits 6-7 of x[0]
	PublicKey     string `json:"public_key,omitempty"`    // 22 bytes (separated only), hex
	Hint          int    `json:"hint,omitempty"`          // hint byte (separated only)
	Raw           string `json:"raw,omitempty"`           // hex of payload when length is unrecognised
}

func processFindMy(data []byte) []byte {
	if len(data) < 1 {
		type Apple struct {
			FindMy `json:"findmy"`
		}
		return mf.MarshalOrEmpty(Apple{FindMy{Variant: "raw"}})
	}

	fm := FindMy{
		Status:       int(data[0]),
		Maintained:   data[0]&0x04 != 0,
		BatteryLevel: findMyBatteryLevel(data[0] >> 6),
	}

	switch len(data) {
	case 2:
		fm.Variant = "nearby"
		fm.PublicKeyBits = int(data[1]) & 0x03
	case 25:
		fm.Variant = "separated"
		fm.PublicKey = hex.EncodeToString(data[1:23])
		fm.PublicKeyBits = int(data[23]) & 0x03
		fm.Hint = int(data[24])
	default:
		fm.Variant = "raw"
		fm.Raw = hex.EncodeToString(data)
	}

	type Apple struct {
		FindMy `json:"findmy"`
	}
	return mf.MarshalOrEmpty(Apple{fm})
}

func findMyBatteryLevel(b byte) string {
	switch b & 0x03 {
	case 0:
		return "full"
	case 1:
		return "medium"
	case 2:
		return "low"
	case 3:
		return "critical"
	}
	return ""
}
