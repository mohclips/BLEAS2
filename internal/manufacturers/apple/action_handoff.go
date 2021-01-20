package apple

import (
	"encoding/binary"
	"encoding/json"

	//"log"
	log "github.com/mohclips/BLEAS2/internal/logging"
)

func processHandoff(data []byte) []byte {

	clipboardStatus := false
	if data[0] != 1 {
		clipboardStatus = true
	}

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

	var mpkt []byte
	var err error
	mpkt, err = json.Marshal(pkt)
	if err != nil {
		log.Error("%s", err)
		mpkt = nil
	}

	return mpkt
}
