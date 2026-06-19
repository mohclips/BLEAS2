package apple

import (
	"encoding/binary"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// TetheringSource decodes the Tethering Source advertisement (subtype 0x0E,
// historically called "hotspot"). Reports battery and cellular signal info
// from a device offering its hotspot to nearby Apple devices.
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/tethering_source.md
type TetheringSource struct {
	Version      int    `json:"version"`
	Flags        int    `json:"flags"`
	BatteryPct   int    `json:"battery_pct"`
	CellType     int    `json:"cell_type"`
	CellTypeName string `json:"cell_type_name"`
	CellBars     int    `json:"cell_bars"`
}

func processTetheringSource(data []byte) []byte {
	type Apple struct {
		TetheringSource `json:"tethering_source"`
	}

	if len(data) < 6 {
		return mf.MarshalOrEmpty(Apple{TetheringSource{}})
	}

	ts := TetheringSource{
		Version:    int(data[0]),
		Flags:      int(data[1]),
		BatteryPct: int(data[2]),
		CellType:   int(binary.LittleEndian.Uint16(data[3:5])),
		CellBars:   int(data[5]),
	}
	if name, ok := cellularTypes[ts.CellType]; ok {
		ts.CellTypeName = name
	}
	return mf.MarshalOrEmpty(Apple{ts})
}

var cellularTypes = map[int]string{
	0: "4G (GSM)",
	1: "1xRTT",
	2: "GPRS",
	3: "EDGE",
	4: "3G (EV-DO)",
	5: "3G",
	6: "4G",
	7: "LTE",
}
