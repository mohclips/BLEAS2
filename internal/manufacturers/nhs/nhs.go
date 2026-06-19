package nhs

import (
	"encoding/hex"
	"fmt"

	log "github.com/mohclips/BLEAS2/internal/logging"
	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
	"github.com/mohclips/BLEAS2/internal/utils"
)

// Parse - parse service data
func Parse(rawData []byte) string {

	// 02 — Subevent code for the HCI_LE_Advertising_Report event
	// 01 — Num of reports
	// 00 — Event type, Connectable and scannable undirected advertising (ADV_IND)
	// 01 — Address type is random device address
	// 12 34 56 78 90 ab — the address in reverse (ab:90:78:56:34:12)
	// xx — svc data length 0x03
	// type 0x03 - Complete List of 16-bit Service Class UUIDs
	// payload 0x6f, 0xfd
	// svc data length 0x17 (includes type byte)
	// type 0x16 - Service Data
	// payload ...

	svcData := rawData[10:]
	if len(svcData) < 1 {
		log.Error("svc data missing length byte")
		return "{}"
	}
	length := int(svcData[0]) // number of AD-entry bytes following the length byte
	if 1+length > len(svcData) {
		log.Error("svc data truncated: length=%d buf=%d", length, len(svcData))
		return "{}"
	}
	payload := svcData[1 : 1+length]
	payloadExtra := svcData[1+length:]

	log.Debug("extra: %s", utils.FormatHexComma(hex.EncodeToString(payloadExtra)))

	// svcsList - List of services contained in this BLE packet
	svcsList := utils.WalkSvcs(payload)

	log.Trace("NHS SvcsList: %+v", svcsList)

	// Now do something with the information returned in the svcsList
	// We should know what the venodr is doing here so we can

	// Build a config map:
	confMap := map[int]utils.Svc{}
	for _, v := range svcsList {
		confMap[v.Action] = v
	}

	var (
		RPI      []byte
		metadata []byte
		extra    []byte
	)
	// And then to find values by key:
	if v, ok := confMap[utils.GAP_SERVICE_DATA]; ok {
		// Found
		//Rolling Proximity Identifier
		RPI = v.ParsedPayload[0:16]
		metadata = v.ParsedPayload[16:20]
		extra = v.ParsedPayload[20:]

		log.Debug("RPI:%x, metadata:%x, extra:%x", RPI, metadata, extra)

	} else {
		log.Error("No NHS payload found")
		return "{}"
	}

	// NHS - packet
	type NHS struct {
		RPI          string `json:"rpi"`
		Metadata     string `json:"metadata"`
		Extra        string `json:"extra"`
		PayloadExtra string `json:"payload_extra"`
	}

	pkt := NHS{
		RPI:          fmt.Sprintf("%x", RPI),
		Metadata:     fmt.Sprintf("%x", metadata),
		Extra:        fmt.Sprintf("%x", extra),
		PayloadExtra: fmt.Sprintf("%x", payloadExtra),
	}

	return `{"nhs":` + string(mf.MarshalOrEmpty(pkt)) + `}`
}
