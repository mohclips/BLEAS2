package apple

import (
	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// AirPlaySource decodes the AirPlay Source advertisement (subtype 0x0A), sent
// when a user enters the "select a source" menu in AirPlay. Single byte
// payload, purpose of the value still undocumented.
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/airplay_source.md
type AirPlaySource struct {
	Data int `json:"data"`
}

func processAirPlaySource(data []byte) []byte {
	type Apple struct {
		AirPlaySource `json:"airplay_source"`
	}

	if len(data) < 1 {
		return mf.MarshalOrEmpty(Apple{AirPlaySource{}})
	}

	return mf.MarshalOrEmpty(Apple{AirPlaySource{Data: int(data[0])}})
}
