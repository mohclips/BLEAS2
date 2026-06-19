package apple

import (
	"encoding/hex"
	"fmt"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// NearbyAction decodes "Nearby Action" advertisements (subtype 0x0F) used to
// trigger setup flows between Apple devices (Apple TV setup, WiFi password
// sharing, AirPods pairing, Companion Link, etc.).
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/nearby_action.md
type NearbyAction struct {
	Flags          int    `json:"flags"`
	ActionType     int    `json:"action_type"`
	ActionTypeName string `json:"action_type_name"`
	AuthTag        string `json:"auth_tag,omitempty"`
	Parameters     string `json:"parameters,omitempty"`
}

func processNearbyAction(data []byte) []byte {
	type Apple struct {
		NearbyAction `json:"nearby_action"`
	}

	if len(data) < 2 {
		return mf.MarshalOrEmpty(Apple{NearbyAction{ActionTypeName: "truncated"}})
	}

	na := NearbyAction{
		Flags:      int(data[0]),
		ActionType: int(data[1]),
	}
	if name, ok := nearbyActionTypes[na.ActionType]; ok {
		na.ActionTypeName = name
	} else {
		na.ActionTypeName = fmt.Sprintf("unknown 0x%02x", na.ActionType)
	}
	if len(data) >= 5 {
		na.AuthTag = hex.EncodeToString(data[2:5])
	}
	if len(data) > 5 {
		na.Parameters = hex.EncodeToString(data[5:])
	}
	return mf.MarshalOrEmpty(Apple{na})
}
