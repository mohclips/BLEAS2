package apple

import (
	"encoding/binary"
	"encoding/hex"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// HeySiri decodes the "Hey Siri" advertisement (subtype 0x08). Sent when a
// device hears the wake phrase; the perceptual hash plus device class are
// broadcast so nearby devices can arbitrate which one responds.
//
// See: https://github.com/furiousMAC/continuity/blob/master/messages/hey_siri.md
type HeySiri struct {
	PerceptualHash  string `json:"perceptual_hash"`
	SNR             int    `json:"snr"`
	Confidence      int    `json:"confidence"`
	DeviceClass     int    `json:"device_class"`
	DeviceClassName string `json:"device_class_name,omitempty"`
	Random          int    `json:"random"`
}

func processHeySiri(data []byte) []byte {
	type Apple struct {
		HeySiri `json:"hey_siri"`
	}

	if len(data) < 7 {
		return mf.MarshalOrEmpty(Apple{HeySiri{}})
	}

	hs := HeySiri{
		PerceptualHash: hex.EncodeToString(data[0:2]),
		SNR:            int(int8(data[2])), // signed dBFS-ish
		Confidence:     int(data[3]),
		DeviceClass:    int(binary.LittleEndian.Uint16(data[4:6])),
		Random:         int(data[6]),
	}
	if name, ok := heySiriDevices[uint16(hs.DeviceClass)]; ok {
		hs.DeviceClassName = name
	}
	return mf.MarshalOrEmpty(Apple{hs})
}
