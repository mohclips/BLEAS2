package apple

import (
	"encoding/binary"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

func processHandoff(data []byte) []byte {

	clipboardStatus := data[0] == 1

	seqNum := binary.BigEndian.Uint16(data[1:3])
	gcmAuth := int(data[3])
	payload := data[4:]

	type Apple struct {
		Handoff `json:"handoff"`
	}

	pkt := Apple{
		Handoff{
			Clipboard: clipboardStatus,
			SeqNum:    seqNum,
			GcmAuth:   gcmAuth,
			Payload:   payload,
		},
	}

	return mf.MarshalOrEmpty(pkt)
}
