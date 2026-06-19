package apple

import (
	"fmt"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// AirPlayTarget decodes the AirPlay Target advertisement (subtype 0x09), sent
// by devices advertising as a receiver (e.g. Apple TV, HomePod).
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/airplay_target.md
type AirPlayTarget struct {
	Flags int    `json:"flags"`
	Seed  int    `json:"seed"`
	IPv4  string `json:"ipv4"`
}

func processAirPlayTarget(data []byte) []byte {
	type Apple struct {
		AirPlayTarget `json:"airplay_target"`
	}

	if len(data) < 6 {
		return mf.MarshalOrEmpty(Apple{AirPlayTarget{}})
	}

	return mf.MarshalOrEmpty(Apple{AirPlayTarget{
		Flags: int(data[0]),
		Seed:  int(data[1]),
		IPv4:  fmt.Sprintf("%d.%d.%d.%d", data[2], data[3], data[4], data[5]),
	}})
}
