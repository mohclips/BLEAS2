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

	GAP_FLAGS                  = 0x01
	GAP_UUID_16BIT_INCOMPLETE  = 0x02
	GAP_UUID_16BIT_COMPLETE    = 0x03
	GAP_UUID_32BIT_INCOMPLETE  = 0x04
	GAP_UUID_32BIT_COMPLETE    = 0x05
	GAP_UUID_128BIT_INCOMPLETE = 0x06
	GAP_UUID_128BIT_COMPLETE   = 0x07
	GAP_NAME_INCOMPLETE        = 0x08
	GAP_NAME_COMPLETE          = 0x09
	GAP_TX_POWER               = 0x0A
	// GAP_ = 0x0B   # UNUSED
	// GAP_ = 0x0C   # UNUSED
	GAP_DEVICE_CLASS                    = 0x0D
	GAP_SIMPLE_PAIRING_HASH             = 0x0E // C or C-192
	GAP_SIMPLE_PAIRING_RANDOMIZER       = 0x0F // R or R-192
	GAP_DEVICE_ID                       = 0x10 // Not Unique: GAP_DEVICE_ID or GAP_SECURITY_MGR_TK_VALUE
	GAP_SECURITY_MGR_OOBand_FLAGS       = 0x11
	GAP_SLAVE_CONNECTION_INTERVAL_RANGE = 0x12
	//  = 0x13   # UNUSED
	GAP_LIST_16BIT_SOLICITATION_UUID   = 0x14
	GAP_LIST_128BIT_SOLICITATION_UUID  = 0x15
	GAP_SERVICE_DATA                   = 0x16
	GAP_PUBLIC_TARGET_ADDRESS          = 0x17
	GAP_RANDOM_TARGET_ADDRESS          = 0x18
	GAP_APPEARANCE                     = 0x19
	GAP_ADVERTISING_INTERVAL           = 0x1A
	GAP_LE_BLUETOOTH_DEVICE_ADDRESS    = 0x1B
	GAP_LE_ROLE                        = 0x1C
	GAP_SIMPLE_PAIRING_HASH_C256       = 0x1D
	GAP_SIMPLE_PAIRING_RANDOMIZER_R256 = 0x1E
	GAP_LIST_32BIT_SOLICITATION_UUID   = 0x1F
	GAP_SERVICE_DATA_32BIT_UUID        = 0x20
	GAP_SERVICE_DATA_128BIT_UUID       = 0x21
	GAP_SECURE_CONN_CONFIRMATION_VAL   = 0x22
	GAP_SECURE_CONN_RANDOM_VAL         = 0x23
	GAP_URI                            = 0x24
	GAP_INDOOR_POSITIONING             = 0x25
	GAP_TRANS_DISC_DATA                = 0x26
	GAP_LE_SUPPORTED_FEATURES          = 0x27
	GAP_CHANNEL_MAP_UPDATE_IND         = 0x28
	GAP_PB_ADV                         = 0x29
	GAP_MESH_MESSAGE                   = 0x2A
	GAP_MESH_BEACON                    = 0x2B
	GAP_BIG_INFO                       = 0x2C
	GAP_BROADCAST_CODE                 = 0x2D
	GAP_3D_INFO_DATA                   = 0x3D
	GAP_MFG_DATA                       = 0xFF
)

var knowniBeaconUUIDs = map[string]string{
	"e2c56db5-dffb-48d2-b060-d0f5a71096e0": "Apple Air Locate (generic uuid)",
	"f7826da6-4fa2-4e98-8024-bc5b71e0893e": "Kontakt.io",
	"2f234454-cf6d-4a0f-adf2-f4911ba9ffa6": "Radius Networks",
	"b9407f30-f5f8-466e-aff9-25556b57fe6d": "Estimote",
	"50765cb7-d9ea-4e21-99a4-fa879613a492": "Common but unknown",
}

// ###########################################################################################################

func LookupiBeaconVendor(uuid string) string {
	var v string
	var found bool
	if v, found = knowniBeaconUUIDs[uuid]; !found {
		v = "unknown"
	}
	return v
}

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
		dest.ParsedPayloadS = ToUUID16(payload)
	case 0x03:
		//Complete List of 16-bit Service Class UUIDs
		dest.ParsedPayloadS = ToUUID16(payload)
	case 0x04:
		//Incomplete List of 32-bit Service Class UUIDs
		dest.ParsedPayloadS = ToUUID32(payload)
	case 0x05:
		//Complete List of 32-bit Service Class UUIDs
		dest.ParsedPayloadS = ToUUID32(payload)
	case 0x06:
		//Incomplete List of 128-bit Service Class UUIDs
		dest.ParsedPayloadS = ToUUID128(payload)
	case 0x07:
		//Complete List of 128-bit Service Class UUIDs
		dest.ParsedPayloadS = ToUUID128(payload)
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
		dest.ParsedPayloadS = ToUUID16(payload)
	case 0x1F:
		//List of 32-bit Service Solicitation UUIDs
		dest.ParsedPayloadS = ToUUID32(payload)
	case 0x15:
		//List of 128-bit Service Solicitation UUIDs
		dest.ParsedPayloadS = ToUUID128(payload)
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
		dest.ParsedPayloadS = ToUUID32(payload)
	case 0x21:
		//Service Data - 128-bit UUID
		dest.ParsedPayloadS = ToUUID128(payload)
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

func ToUUID16(u *[]byte) string {
	//return "0000-" + string((*u)[1]+(*u)[0]) + BaseUUID
	//FIXME: add length check
	return fmt.Sprintf("0000-%x%x%s", (*u)[1], (*u)[0], BaseUUID)
}

func ToUUID32(u *[]byte) string {
	//return string((*u)[1]+(*u)[0]) + "-" + string((*u)[4]+(*u)[3]) + BaseUUID
	//FIXME: add length check
	return fmt.Sprintf("%x%x-%x%x%s", (*u)[1], (*u)[0], (*u)[3], (*u)[2], BaseUUID)
}

func ToUUID128(u *[]byte) string {

	//FIXME: add length check
	//return fmt.Sprintf(GenericUUID128, (*u)[1], (*u)[0], (*u)[4], (*u)[3])
	return "TBC"
}

func ToUUID128iBeacon(u *[]byte) string {

	//FIXME: add length check
	return fmt.Sprintf(GenericUUID128, (*u)[0], (*u)[1], (*u)[2], (*u)[3],
		(*u)[4], (*u)[5], (*u)[6], (*u)[7],
		(*u)[8], (*u)[9], (*u)[10], (*u)[11],
		(*u)[12], (*u)[13], (*u)[14], (*u)[15])
}
