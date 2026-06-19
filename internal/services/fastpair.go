package services

import (
	"encoding/hex"
	"fmt"
)

// FastPair is Google's proximity-pairing protocol. The service data has two
// distinct forms:
//
//   - **Discoverable**: 3-byte payload = Model ID (24-bit). The receiver
//     can look this up in the Fast Pair model database to identify the
//     exact accessory (e.g. Pixel Buds A-Series, Bose QC45, JBL Live Pro).
//   - **Not discoverable**: leading 1-byte flags + variable account-key
//     filter, used when the device is already paired and just announcing
//     presence.
//
// Spec: https://developers.google.com/nearby/fast-pair/specifications/service/provider
type FastPair struct {
	UUID         string `json:"uuid"`                     // "fef3" or "fe2c"
	Mode         string `json:"mode"`                     // "discoverable" | "not_discoverable" | "unknown"
	ModelID      string `json:"model_id,omitempty"`       // 24-bit hex (discoverable mode)
	Flags        int    `json:"flags,omitempty"`          // non-discoverable header byte
	AccountKey   string `json:"account_key_filter,omitempty"`
	Raw          string `json:"raw,omitempty"`
}

func parseFastPair(data []byte, uuid string) []byte {
	fp := FastPair{UUID: uuid}
	switch {
	case len(data) == 3:
		// Discoverable: 24-bit big-endian Model ID.
		fp.Mode = "discoverable"
		fp.ModelID = fmt.Sprintf("%02x%02x%02x", data[0], data[1], data[2])
	case len(data) >= 1:
		// Not-discoverable: leading header byte then opaque filter.
		fp.Mode = "not_discoverable"
		fp.Flags = int(data[0])
		if len(data) > 1 {
			fp.AccountKey = hex.EncodeToString(data[1:])
		}
	default:
		fp.Mode = "unknown"
		fp.Raw = hex.EncodeToString(data)
	}
	return raw(fp)
}
