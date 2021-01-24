package utils

// routines to parse BLE advertisment services

import (
	"fmt"

	log "github.com/mohclips/BLEAS2/internal/logging"
)

// Svc - parsed service record
// allows for string returned or a byte slice
type Svc struct {
	Action         int
	ParsedPayload  []byte
	ParsedPayloadS string
}

// 0000xxxx-0000-1000-8000-00805F9B34FB
// xxxxxxxx-0000-1000-8000-00805F9B34FB
// 128_bit_value = 16_bit_value * 2^96 + Bluetooth_Base_UUID
// 128_bit_value = 32_bit_value * 2^96 + Bluetooth_Base_UUID

// https://github.com/hbldh/bleak/blob/develop/bleak/uuids.py

const (
	// BaseUUID ...
	BaseUUID string = "-0000-1000-8000-00805F9B34FB"
	// BaseUUID16 ...
	BaseUUID16 string = "0000xxxx-0000-1000-8000-00805F9B34FB"
	// BaseUUID32 ...
	BaseUUID32 string = "xxxxxxxx-0000-1000-8000-00805F9B34FB"
	// GenericUUID128 ...
	GenericUUID128 string = "%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x"
)

// ###########################################################################################################

// WalkSvcs - walk though the Advertisement Services
// and update a slice with parsed details
func WalkSvcs(svcs *[]byte, svcsList *[]Svc) {

	allSvcsLen := int((*svcs)[0:][0]) // How big is this Services payload

	var (
		thisSvcLen     int
		thisSvcType    int
		thisSvcPayload []byte
		i              int = 1 // skip first byte which is the full length
	)

	for i < allSvcsLen {

		// grab first two bytes (length ,type)
		thisSvcLen = int((*svcs)[i:][0]) - 1 // Length Byte (-1 to take care of the Type byte)
		i++                                  // move to next byte
		thisSvcType = int((*svcs)[i:][0])    // Service Type Byte - determines what to do with the payload data
		i++                                  // move to next byte
		// grab payload
		thisSvcPayload = (*svcs)[i:][0:thisSvcLen]

		payloadLen := len(thisSvcPayload)

		if payloadLen != thisSvcLen {
			log.Error("wrong payload length")
		}

		//log.Info("thisSvcLen: %d, thisSvcType:0x%02x", thisSvcLen, thisSvcType)
		//log.Info("thisSvcLen: 0x%02x, thisSvcType:0x%02x, payload_len: 0x%02x, payload %x", thisSvcLen, thisSvcType, len(thisSvcPayload), thisSvcPayload)

		// do something with this service

		log.Trace("do Action: %d, %x, %x", thisSvcLen, thisSvcType, thisSvcPayload)

		// svc := Svcs{
		// 	Action:        thisSvcType,
		// 	ParsedPayload: s,
		// }

		var parsedSvc Svc

		doSvcAction(thisSvcType, &thisSvcPayload, &parsedSvc)

		*svcsList = append(*svcsList, parsedSvc)

		log.Debug("%+v", parsedSvc)

		// move to the next service
		i = i + thisSvcLen

		//var svc Svcs

	}

}

// DoSvcAction - parse the service payload depending on action type
func doSvcAction(action int, payload *[]byte, dest *Svc) {

	dest.Action = action
	switch action {
	case 0x01:
		//Flags
		dest.ParsedPayloadS = string(*payload)
	case 0x02:
		//Incomplete List of 16-bit Service Class UUIDs
		dest.ParsedPayloadS = toUUID16(payload)
	case 0x03:
		//Complete List of 16-bit Service Class UUIDs
		dest.ParsedPayloadS = toUUID16(payload)
	case 0x04:
		//Incomplete List of 32-bit Service Class UUIDs
		dest.ParsedPayloadS = toUUID32(payload)
	case 0x05:
		//Complete List of 32-bit Service Class UUIDs
		dest.ParsedPayloadS = toUUID32(payload)
	case 0x06:
		//Incomplete List of 128-bit Service Class UUIDs
		dest.ParsedPayloadS = toUUID128(payload)
	case 0x07:
		//Complete List of 128-bit Service Class UUIDs
		dest.ParsedPayloadS = toUUID128(payload)
	case 0x08:
		//Shortened Local Name
		dest.ParsedPayloadS = string(*payload)
	case 0x09:
		//Complete Local Name
		dest.ParsedPayloadS = string(*payload)
	case 0x0A:
		//Tx Power Level
		dest.ParsedPayloadS = fmt.Sprintf("%x", *payload) //FIXME: Needs attention
	case 0x0D:
		//Class of Device
		dest.ParsedPayloadS = fmt.Sprintf("%x", *payload) //FIXME: Needs attention
	case 0x0E:
		//Simple Pairing Hash C
		dest.ParsedPayloadS = fmt.Sprintf("%x", *payload) //FIXME: Needs attention
	case 0x0F:
		//Simple Pairing Randomizer R
		dest.ParsedPayloadS = fmt.Sprintf("%x", *payload) //FIXME: Needs attention
	case 0x10:
		//Device ID
		dest.ParsedPayloadS = fmt.Sprintf("%x", *payload) //FIXME: Needs attention
	// case 0x10 : //Security Manager TK Value
	case 0x11:
		//Security Manager Out of Band Flags
		dest.ParsedPayloadS = fmt.Sprintf("%x", *payload) //FIXME: Needs attention
	case 0x12:
		//Slave Connection Interval Range
		dest.ParsedPayloadS = fmt.Sprintf("%x", *payload) //FIXME: Needs attention
	case 0x14:
		//List of 16-bit Service Solicitation UUIDs
		dest.ParsedPayloadS = toUUID16(payload)
	case 0x1F:
		//List of 32-bit Service Solicitation UUIDs
		dest.ParsedPayloadS = toUUID32(payload)
	case 0x15:
		//List of 128-bit Service Solicitation UUIDs
		dest.ParsedPayloadS = toUUID128(payload)
	case 0x16:
		//Service Data
		// return fmt.Sprintf("%x", *payload)
		dest.ParsedPayload = *payload
	case 0x17:
		//Public Target Address
		//return fmt.Sprintf("%x", *payload)
		dest.ParsedPayload = *payload
	case 0x18:
		//Random Target Address
		//return fmt.Sprintf("%x", *payload)
		dest.ParsedPayload = *payload
	case 0x19:
		//Appearance
		dest.ParsedPayload = *payload
	case 0x1A:
		//Advertising Interval
		//return fmt.Sprintf("%x", *payload)
		dest.ParsedPayload = *payload
	case 0x1B:
		//LE Bluetooth Device Address
		//return fmt.Sprintf("%x", *payload)
		dest.ParsedPayload = *payload
	case 0x1C:
		//LE Role
		//return ""
		dest.ParsedPayload = *payload
	case 0x1D:
		//Simple Pairing Hash C-256
		//return fmt.Sprintf("%x", *payload)
		dest.ParsedPayload = *payload
	case 0x1E:
		//Simple Pairing Randomizer R-256
		//return fmt.Sprintf("%x", *payload)
		dest.ParsedPayload = *payload
	case 0x20:
		//Service Data - 32-bit UUID
		dest.ParsedPayloadS = toUUID32(payload)
	case 0x21:
		//Service Data - 128-bit UUID
		dest.ParsedPayloadS = toUUID128(payload)
	case 0x3D:
		//3D Information Data
		//return ""
		dest.ParsedPayload = *payload
	case 0xFF:
		//Manufacturer Specific Data
		//return ""
		dest.ParsedPayload = *payload

	}

}

func toUUID16(u *[]byte) string {
	//return "0000-" + string((*u)[1]+(*u)[0]) + BaseUUID
	return fmt.Sprintf("0000-%x%x%s", (*u)[1], (*u)[0], BaseUUID)
}

func toUUID32(u *[]byte) string {
	//return string((*u)[1]+(*u)[0]) + "-" + string((*u)[4]+(*u)[3]) + BaseUUID
	return fmt.Sprintf("%x%x-%x%x%s", (*u)[1], (*u)[0], (*u)[4], (*u)[3], BaseUUID)
}

func toUUID128(u *[]byte) string {

	//return fmt.Sprintf("%s", GenericUUID128, (*u)[1], (*u)[0], (*u)[4], (*u)[3])
	return "TBC"
}
