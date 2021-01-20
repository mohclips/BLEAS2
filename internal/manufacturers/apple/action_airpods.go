package apple

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	//log"
	log "github.com/mohclips/BLEAS2/internal/logging"
)

//     # https://github.com/furiousMAC/continuity/blob/master/messages/proximity_pairing.md

func processAirpods(data []byte) []byte {

	rawUndef1 := int(data[0]) // 0x01
	rawDeviceModel := binary.BigEndian.Uint16(data[1:3])
	rawStatus := int(data[3])
	rawBatteryRL := int(data[4])
	rawPower := int(data[5]) // '? C R L xxxx' xxxx = case battery
	rawLid := int(data[6])
	rawColor := int(data[7])
	rawUndef2 := int(data[8]) // 0x00
	rawPayload := data[9:]

	batteryR := int(data[4]) & 0b11110000 >> 4
	batteryL := int(data[4]) & 0b00001111

	C := int(data[5]&0b01000000) >> 7
	R := int(data[5]&0b00100000) >> 6
	L := int(data[5]&0b00010000) >> 5
	casePower := int(data[5]) & 0b00001111

	var deviceModel string
	var found bool
	if deviceModel, found = airpodDevices[rawDeviceModel]; !found {
		deviceModel = fmt.Sprintf("unknown model: 0x%02x", rawDeviceModel)
	}

	//fmt.Printf("\ndm: %02x   %s\n", rawDeviceModel, deviceModel)

	type Apple struct {
		Airpods `json:"airpods"`
	}

	pkt := Apple{
		Airpods{
			RawUndef1:      rawUndef1,
			RawDeviceModel: rawDeviceModel,
			DeviceModel:    deviceModel,
			RawStatus:      rawStatus,
			RawBatteryRL:   rawBatteryRL,
			Battery: sBattery{
				R: batteryR,
				L: batteryL,
			},
			RawPower: rawPower,
			Charging: sCharging{
				Case: C,
				R:    R,
				L:    L,
			},
			CasePower:  casePower,
			Lid:        rawLid,
			RawColor:   rawColor,
			RawUndef2:  rawUndef2,
			RawPayload: rawPayload,
		},
	}

	// convert to json
	var mpkt []byte
	var err error
	mpkt, err = json.Marshal(pkt)
	if err != nil {
		log.Error("%s", err)
		mpkt = nil
	}

	return mpkt
}
