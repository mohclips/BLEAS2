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

	log.Debug("RAW: %s", utils.FormatHexComma(hex.EncodeToString(rawData)))
	log.Debug("subevent:%d reports:%d", rawData[0], rawData[1])
	log.Debug("eventType: %d", eventType)
	log.Debug("addressType: %d", rawData[3])
	log.Debug("SVC: %s", utils.FormatHexComma(hex.EncodeToString(svcData)))
	log.Debug("payload: %s", utils.FormatHexComma(hex.EncodeToString(payload)))
	log.Debug("all svc len: %x %d", length, length)

	// .. more bytes in pkt?, then get length,get type, parse...  NEEDS pointers, to rawData to check pointer adddress against legth of rawdata
	log.Debug("extra: %s", utils.FormatHexComma(hex.EncodeToString(extra))) // what is this?

	walkSvcs(&payload)

	return ""
}

func walkSvcs(svcs *[]byte) string {

	allSvcsLen := int((*svcs)[0:][0])

	log.Info("allSvcsLen: 0x%02x ", allSvcsLen) // length of the whole lot

	var (
		newSvc          bool = true
		thisSvcLen      int
		thisSvcLenSaved int
		thisSvcType     int
		thisSvcPayload  []byte
		i               int = 1 // skip first byte which is the full length
	)

	for i < allSvcsLen {

		sByte := (*svcs)[i:][0] // a byte in the list
		//log.Info("> %d %x", i, sByte)

		if newSvc {
			// grab first two bytes
			thisSvcLen = int((*svcs)[i:][0])  // Length Byte - decrement as we move thru the array/slice
			thisSvcLenSaved = thisSvcLen      // save this for later
			i++                               // move to next byte
			thisSvcType = int((*svcs)[i:][0]) // Service Type Byte - determines what to do with the payload data

			//log.Info("thisSvcLen: %d, thisSvcType:0x%02x", thisSvcLen, thisSvcType)

			newSvc = false
			thisSvcLen--
		} else {
			// rest of data is payload
			thisSvcPayload = append(thisSvcPayload, sByte)
			// until end of thisSvcLen
			thisSvcLen--
		}

		i++ // move to next byte

		if thisSvcLen == 0 || i == allSvcsLen {
			//log.Info("Done: %d, %d, %x", thisSvcLen, thisSvcType, thisSvcPayload)

			// do something with this service
			log.Info("thisSvcLen: 0x%02x, thisSvcType:0x%02x, payload_len: 0x%02x, payload %x", thisSvcLenSaved, thisSvcType, len(thisSvcPayload), thisSvcPayload)

			// clean up
			thisSvcPayload = nil

			// next byte is a new Service
			newSvc = true
		}

	}

	return ""
}

func toOctetStringArray(data *[]byte, num int) string {
	return "boom"
}

func toStringArray() {

}

func toString() {

}
func toSignedInt() {

}

// var ServiceFuncMap = map[int]serviceFuncStruct{
// 	0x01: {name: "Flags", resolve: toStringArray()},
// 	0x02: {name: "Incomplete List of 16-bit Service Class UUIDs", resolve: toOctetStringArrayTwo},
// 	0x03: {name: "Complete List of 16-bit Service Class UUIDs", resolve: toOctetStringArray(nil, 2)},
// 	0x04: {name: "Incomplete List of 32-bit Service Class UUIDs", resolve: toOctetStringArray(nil, 4)},
// 	0x05: {name: "Complete List of 32-bit Service Class UUIDs", resolve: toOctetStringArray(nil, 4)},
// 	0x06: {name: "Incomplete List of 128-bit Service Class UUIDs", resolve: toOctetStringArray(nil, 16)},
// 	0x07: {name: "Complete List of 128-bit Service Class UUIDs", resolve: toOctetStringArray(nil, 16)},
// 	0x08: {name: "Shortened Local Name", resolve: toString()},
// 	0x09: {name: "Complete Local Name", resolve: toString()},
// 	0x0A: {name: "Tx Power Level", resolve: toSignedInt()},
// 	0x0D: {name: "Class of Device", resolve: toOctetString(nil, 3)},
// 	0x0E: {name: "Simple Pairing Hash C", resolve: toOctetString(nil, 16)},
// 	0x0F: {name: "Simple Pairing Randomizer R", resolve: toOctetString(nil, 16)},
// 	0x10: {name: "Device ID", resolve: toOctetString(nil, 16)},
// 	// 0x10 : { name : "Security Manager TK Value", resolve: nil }
// 	0x11: {name: "Security Manager Out of Band Flags", resolve: toOctetString(nil, 16)},
// 	0x12: {name: "Slave Connection Interval Range", resolve: toOctetStringArray(nil, 2)},
// 	0x14: {name: "List of 16-bit Service Solicitation UUIDs", resolve: toOctetStringArray(nil, 2)},
// 	0x1F: {name: "List of 32-bit Service Solicitation UUIDs", resolve: toOctetStringArray(nil, 4)},
// 	0x15: {name: "List of 128-bit Service Solicitation UUIDs", resolve: toOctetStringArray(nil, 8)},
// 	0x16: {name: "Service Data", resolve: toOctetStringArray(nil, 1)},
// 	0x17: {name: "Public Target Address", resolve: toOctetStringArray(nil, 6)},
// 	0x18: {name: "Random Target Address", resolve: toOctetStringArray(nil, 6)},
// 	0x19: {name: "Appearance", resolve: nil},
// 	0x1A: {name: "Advertising Interval", resolve: toOctetStringArray(nil, 2)},
// 	0x1B: {name: "LE Bluetooth Device Address", resolve: toOctetStringArray(nil, 6)},
// 	0x1C: {name: "LE Role", resolve: nil},
// 	0x1D: {name: "Simple Pairing Hash C-256", resolve: toOctetStringArray(nil, 16)},
// 	0x1E: {name: "Simple Pairing Randomizer R-256", resolve: toOctetStringArray(nil, 16)},
// 	0x20: {name: "Service Data - 32-bit UUID", resolve: toOctetStringArray(nil, 4)},
// 	0x21: {name: "Service Data - 128-bit UUID", resolve: toOctetStringArray(nil, 16)},
// 	0x3D: {name: "3D Information Data", resolve: nil},
// 	0xFF: {name: "Manufacturer Specific Data", resolve: nil},
// }
