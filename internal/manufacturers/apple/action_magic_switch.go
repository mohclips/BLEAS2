package apple

import (
	"encoding/hex"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// MagicSwitch decodes the Magic Switch / watch-on-wrist advertisement
// (subtype 0x0B). Sent by an Apple Watch whose screen is on but has lost its
// paired iPhone connection.
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/magic_switch.md
type MagicSwitch struct {
	Data       string `json:"data"`
	Confidence int    `json:"confidence"`
	OnWrist    bool   `json:"on_wrist"`
}

func processMagicSwitch(data []byte) []byte {
	type Apple struct {
		MagicSwitch `json:"magic_switch"`
	}

	if len(data) < 3 {
		return mf.MarshalOrEmpty(Apple{MagicSwitch{}})
	}

	return mf.MarshalOrEmpty(Apple{MagicSwitch{
		Data:       hex.EncodeToString(data[0:2]),
		Confidence: int(data[2]),
		OnWrist:    data[2] == 0x3F,
	}})
}
