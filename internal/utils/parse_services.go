package utils

// routines to parse BLE advertisment services

import (
	"fmt"

	log "github.com/mohclips/BLEAS2/internal/logging"
)

// Svc - parsed service record (string OR raw bytes form, depending on action).
type Svc struct {
	Action         int
	ParsedPayload  []byte
	ParsedPayloadS string
}

// https://github.com/hbldh/bleak/blob/develop/bleak/uuids.py
const (
	BaseUUID       = "-0000-1000-8000-00805F9B34FB"
	GenericUUID128 = "%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x"

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
	GAP_DEVICE_CLASS           = 0x0D
	GAP_SIMPLE_PAIRING_HASH    = 0x0E
	GAP_SIMPLE_PAIRING_RANDOMIZER       = 0x0F
	GAP_DEVICE_ID                       = 0x10
	GAP_SECURITY_MGR_OOBand_FLAGS       = 0x11
	GAP_SLAVE_CONNECTION_INTERVAL_RANGE = 0x12
	GAP_LIST_16BIT_SOLICITATION_UUID    = 0x14
	GAP_LIST_128BIT_SOLICITATION_UUID   = 0x15
	GAP_SERVICE_DATA                    = 0x16
	GAP_PUBLIC_TARGET_ADDRESS           = 0x17
	GAP_RANDOM_TARGET_ADDRESS           = 0x18
	GAP_APPEARANCE                      = 0x19
	GAP_ADVERTISING_INTERVAL            = 0x1A
	GAP_LE_BLUETOOTH_DEVICE_ADDRESS     = 0x1B
	GAP_LE_ROLE                         = 0x1C
	GAP_SIMPLE_PAIRING_HASH_C256        = 0x1D
	GAP_SIMPLE_PAIRING_RANDOMIZER_R256  = 0x1E
	GAP_LIST_32BIT_SOLICITATION_UUID    = 0x1F
	GAP_SERVICE_DATA_32BIT_UUID         = 0x20
	GAP_SERVICE_DATA_128BIT_UUID        = 0x21
	GAP_SECURE_CONN_CONFIRMATION_VAL    = 0x22
	GAP_SECURE_CONN_RANDOM_VAL          = 0x23
	GAP_URI                             = 0x24
	GAP_INDOOR_POSITIONING              = 0x25
	GAP_TRANS_DISC_DATA                 = 0x26
	GAP_LE_SUPPORTED_FEATURES           = 0x27
	GAP_CHANNEL_MAP_UPDATE_IND          = 0x28
	GAP_PB_ADV                          = 0x29
	GAP_MESH_MESSAGE                    = 0x2A
	GAP_MESH_BEACON                     = 0x2B
	GAP_BIG_INFO                        = 0x2C
	GAP_BROADCAST_CODE                  = 0x2D
	GAP_3D_INFO_DATA                    = 0x3D
	GAP_MFG_DATA                        = 0xFF
)

var knowniBeaconUUIDs = map[string]string{
	"e2c56db5-dffb-48d2-b060-d0f5a71096e0": "Apple Air Locate (generic uuid)",
	"f7826da6-4fa2-4e98-8024-bc5b71e0893e": "Kontakt.io",
	"2f234454-cf6d-4a0f-adf2-f4911ba9ffa6": "Radius Networks",
	"b9407f30-f5f8-466e-aff9-25556b57fe6d": "Estimote",
	"50765cb7-d9ea-4e21-99a4-fa879613a492": "Common but unknown",
}

func LookupiBeaconVendor(uuid string) string {
	if v, ok := knowniBeaconUUIDs[uuid]; ok {
		return v
	}
	return "unknown"
}

// WalkSvcs walks a sequence of AD entries (`[length, type, data...]`...) and
// returns parsed records. The input must contain only the entries — no leading
// total-length byte.
func WalkSvcs(svcs []byte) []Svc {
	var out []Svc
	i := 0
	for i < len(svcs) {
		if i+2 > len(svcs) {
			log.Error("truncated svc header at offset %d (buf %d)", i, len(svcs))
			return out
		}
		thisSvcLen := int(svcs[i]) - 1 // length byte includes the type byte
		i++
		thisSvcType := int(svcs[i])
		i++

		if thisSvcLen < 0 || i+thisSvcLen > len(svcs) {
			log.Error("svc payload exceeds buffer: i=%d len=%d buf=%d", i, thisSvcLen, len(svcs))
			return out
		}
		thisSvcPayload := svcs[i : i+thisSvcLen]

		log.Trace("do Action: %d, %x, %x", thisSvcLen, thisSvcType, thisSvcPayload)
		parsed := doSvcAction(thisSvcType, thisSvcPayload)
		out = append(out, parsed)
		log.Debug("%+v", parsed)

		i += thisSvcLen
	}
	return out
}

// doSvcAction parses one service payload according to its action type.
func doSvcAction(action int, payload []byte) Svc {
	s := Svc{Action: action}
	switch action {
	case 0x01: // Flags
		s.ParsedPayloadS = string(payload)
	case 0x02, 0x03: // 16-bit Service Class UUIDs (Incomplete/Complete)
		s.ParsedPayloadS = ToUUID16(payload)
	case 0x04, 0x05: // 32-bit Service Class UUIDs
		s.ParsedPayloadS = ToUUID32(payload)
	case 0x06, 0x07: // 128-bit Service Class UUIDs
		s.ParsedPayloadS = ToUUID128(payload)
	case 0x08, 0x09: // Local Name (Shortened/Complete)
		s.ParsedPayloadS = string(payload)
	case 0x0A, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12: // hex-dumped scalars/structs
		s.ParsedPayloadS = fmt.Sprintf("%x", payload)
	case 0x14: // List of 16-bit Service Solicitation UUIDs
		s.ParsedPayloadS = ToUUID16(payload)
	case 0x15: // List of 128-bit Service Solicitation UUIDs
		s.ParsedPayloadS = ToUUID128(payload)
	case 0x1F: // List of 32-bit Service Solicitation UUIDs
		s.ParsedPayloadS = ToUUID32(payload)
	case 0x20: // Service Data - 32-bit UUID
		s.ParsedPayloadS = ToUUID32(payload)
	case 0x21: // Service Data - 128-bit UUID
		s.ParsedPayloadS = ToUUID128(payload)
	case 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x3D, 0xFF:
		s.ParsedPayload = payload
	}
	return s
}

func ToUUID16(u []byte) string {
	if len(u) < 2 {
		return ""
	}
	return fmt.Sprintf("0000%02x%02x%s", u[1], u[0], BaseUUID)
}

func ToUUID32(u []byte) string {
	if len(u) < 4 {
		return ""
	}
	return fmt.Sprintf("%02x%02x%02x%02x%s", u[3], u[2], u[1], u[0], BaseUUID)
}

// ToUUID128 - BLE 128-bit UUIDs are transmitted little-endian over the air.
func ToUUID128(u []byte) string {
	if len(u) < 16 {
		return ""
	}
	return fmt.Sprintf(GenericUUID128,
		u[15], u[14], u[13], u[12],
		u[11], u[10], u[9], u[8],
		u[7], u[6], u[5], u[4],
		u[3], u[2], u[1], u[0])
}

// ToUUID128iBeacon - iBeacon UUIDs are big-endian inside the payload.
func ToUUID128iBeacon(u []byte) string {
	if len(u) < 16 {
		return ""
	}
	return fmt.Sprintf(GenericUUID128,
		u[0], u[1], u[2], u[3],
		u[4], u[5], u[6], u[7],
		u[8], u[9], u[10], u[11],
		u[12], u[13], u[14], u[15])
}
