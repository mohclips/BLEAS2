package apple

import (
	"encoding/hex"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// AirDrop decodes AirDrop "looking for receivers" advertisements (subtype 0x05).
// The first 8 bytes are zero padding; then a version, then four SHA-256
// prefixes (Apple ID / phone / email / email2), then a trailing zero byte.
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/airdrop.md
type AirDrop struct {
	Version     int    `json:"version"`
	AppleIDHash string `json:"apple_id_hash"`
	PhoneHash   string `json:"phone_hash"`
	EmailHash   string `json:"email_hash"`
	Email2Hash  string `json:"email2_hash"`
}

func processAirDrop(data []byte) []byte {
	type Apple struct {
		AirDrop `json:"airdrop"`
	}

	// Need at least up to byte 17 (Email2 hash high byte).
	if len(data) < 17 {
		return mf.MarshalOrEmpty(Apple{AirDrop{}})
	}

	return mf.MarshalOrEmpty(Apple{AirDrop{
		Version:     int(data[8]),
		AppleIDHash: hex.EncodeToString(data[9:11]),
		PhoneHash:   hex.EncodeToString(data[11:13]),
		EmailHash:   hex.EncodeToString(data[13:15]),
		Email2Hash:  hex.EncodeToString(data[15:17]),
	}})
}
