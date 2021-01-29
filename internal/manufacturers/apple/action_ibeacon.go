package apple

import (
	"encoding/binary"
	"encoding/json"

	//"log"
	log "github.com/mohclips/BLEAS2/internal/logging"
	"github.com/mohclips/BLEAS2/internal/utils"
)

// https://en.wikipedia.org/wiki/IBeacon#BLE_Advertisement_Packet_Structure_Byte_Map
//  Byte 3: Length:             0x1a (Of the following section)
//  Byte 4: Type:               0xff (Custom Manufacturer Data)
//  Byte 5-6: Manufacturer ID : 0x4c00 (Apple's Bluetooth SIG registered company code, 16-bit Little Endian)
//  Byte 7: SubType:            0x02 (Apple's iBeacon type of Custom Manufacturer Data)

//  Byte 8: SubType Length:     0x15 (Of the rest of the iBeacon data; UUID + Major + Minor + TXPower)
//  Byte 9-24: Proximity UUID        (Random or Public/Registered UUID of the specific beacon)
//  Byte 25-26: Major                (User-Defined value)
//  Byte 27-28: Minor                (User-Defined value)
//  Byte 29: TXPower                 (8 bit Signed value, ranges from -128 to 127, use Two's Compliment to "convert" if necessary, Units: Measured Transmission Power in dBm @ 1 meters from beacon) (Set by user, not dynamic, can be used in conjunction with the received RSSI at a receiver to calculate rough distance to beacon)

func processiBeacon(data []byte) []byte {

	// -1 subtype
	// 0:15 uuid
	// 16:17 major
	// 18:19 minor
	uuid := data[0:16]
	proximityUUID := utils.ToUUID128iBeacon(&uuid)

	name := utils.LookupiBeaconVendor(proximityUUID)

	major := binary.BigEndian.Uint16(data[16:18])
	minor := binary.BigEndian.Uint16(data[18:20])

	type Apple struct {
		iBeacon `json:"ibeacon"`
	}

	pkt := Apple{
		iBeacon{
			UUID:  proximityUUID,
			Name:  name,
			Major: major,
			Minor: minor,
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
