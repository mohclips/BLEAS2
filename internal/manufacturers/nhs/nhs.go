package nhs

import (
	"encoding/hex"

	//"log"
	log "github.com/mohclips/BLEAS2/internal/logging"

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

	// 00 — Event type, Connectable and scannable undirected advertising (ADV_IND)
	eventType := rawData[2]
	svcData := rawData[10:]
	length := svcData[0] // length of all services
	payload := svcData[0:length]
	extra := svcData[length:] // what is this?

	log.Debug("rawRAW: %x", rawData)
	log.Debug("RAW: %s", utils.FormatHexComma(hex.EncodeToString(rawData)))
	log.Debug("subevent:%d reports:%d", rawData[0], rawData[1])
	log.Debug("eventType: %d", eventType)
	log.Debug("addressType: %d", rawData[3])
	log.Debug("SVC: %s", utils.FormatHexComma(hex.EncodeToString(svcData)))
	log.Debug("payload: %s", utils.FormatHexComma(hex.EncodeToString(payload)))
	log.Debug("all svc len: %x %d", length, length)

	// .. more bytes in pkt?, then get length,get type, parse...  NEEDS pointers, to rawData to check pointer adddress against legth of rawdata
	log.Debug("extra: %s", utils.FormatHexComma(hex.EncodeToString(extra))) // what is this?
	//log.Debug("extra signed int: %d", int(extra[0]))

	if len(payload) != int(length) {
		log.Error("wrong payload length: payload: %d  length:%d", len(payload), int(length))
	}

	// svcsList - List of services contained in this BLE packet
	var svcsList []utils.Svc

	utils.WalkSvcs(&payload, &svcsList)

	log.Info("NHS SvcsList: %+v", svcsList)

	// Now do something with the information returned in the svcsList

	return "" // some json
}
