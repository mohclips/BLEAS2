package apple

import (
	"encoding/hex"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// TetheringTarget decodes the Tethering Target advertisement (subtype 0x0D,
// historically called "wifi_set"). Carries a 4-byte iCloud DSID-derived
// identifier that rotates daily.
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/tethering_target.md
type TetheringTarget struct {
	ICloudID string `json:"icloud_id"`
}

func processTetheringTarget(data []byte) []byte {
	type Apple struct {
		TetheringTarget `json:"tethering_target"`
	}

	if len(data) < 4 {
		return mf.MarshalOrEmpty(Apple{TetheringTarget{}})
	}

	return mf.MarshalOrEmpty(Apple{TetheringTarget{
		ICloudID: hex.EncodeToString(data[0:4]),
	}})
}
