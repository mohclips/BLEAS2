package apple

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mohclips/BLEAS2/internal/utils"
)

func parseOsWifiCode(code int, dev string) (string, string) {

	switch sw := code; sw {
	case 0x1c:
		if dev == "MacBook" {
			return "Mac OS", "On"
		}
		return "iOS12", "On"

	case 0x18:
		if dev == "MacBook" {
			return "Mac OS", "Off"
		}
		return "iOS12", "Off"

	case 0x10:
		return "iOS11", "<unknown>"
	case 0x1e:
		return "iOS13", "On"
	case 0x1a:
		return "iOS13", "Off"
	case 0x0e:
		return "iOS13", "Connecting"
	case 0x0c:
		return "iOS12", "On"
	case 0x04:
		return "iOS13", "On"
	case 0x00:
		return "iOS10", "<unknown>"
	case 0x09:
		return "Mac OS", "<unknown>"
	case 0x14:
		return "Mac OS", "On"
	case 0x98:
		return "WatchOS", "<unknown>"
	default:
		return "", ""
	}
}

func processNearby(data []byte) []byte {

	phoneStateID := int(data[0] & 0b00001111)
	var phoneState string
	var found bool
	if phoneState, found = phoneStates[phoneStateID]; !found {
		phoneState = fmt.Sprintf("unknown state: %x", phoneStateID)
	}

	// nearby status mask
	phoneStatusMask := (int(data[0]) & 0b11110000) >> 4
	// pick each bitwise flag from map/dict
	masks := utils.BitmaskToNames(phoneStatusMask, nearbyStatusMasks)

	os, wifi := parseOsWifiCode(int(data[1]), "") //we dont know the device type

	type Apple struct {
		Nearby `json:"nearby"`
	}
	pkt := Apple{
		Nearby{
			State:    phoneState,
			RawState: phoneStateID,
			Masks:    masks,
			RawMask:  phoneStatusMask,
			Wifi:     wifi,
			RawWifi:  int(data[1]),
			Os:       os,
		},
	}

	var mpkt []byte
	var err error
	mpkt, err = json.Marshal(pkt)
	if err != nil {
		log.Println(err)
	}

	//	fmt.Printf("mpkt %s\n", mpkt)

	return mpkt
}
