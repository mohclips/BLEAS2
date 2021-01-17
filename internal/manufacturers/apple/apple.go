package apple

import (
	"encoding/json"
	"fmt"
	"log"
)

////////////////////////////////////////////////////////////////////////////////////////////////

// ParseMF - parse manufacturers data
func ParseMF(mmfData []byte) string {

	var jApple []byte // Apple struct as json

	// mfData contains the Maufacturer ID two bytes at start
	mfData := mmfData[2:]

	action := int(mfData[0])
	length := int(mfData[1])

	if length > len(mfData)-1 || length == 0 {
		fmt.Printf("ERROR: pkt length too short, %d\n", length)
		return ""
	}

	data := mfData[2:length] // the actual packet data

	//log.Printf("Action: 0x%02x Len: %d", action, length)

	if len(data) == 0 {
		return ""
	}

	if len(mfData) > length+2 {
		log.Printf("WARNING: manufacturers data is greater than one action, %+v", mfData[length:])
		//TODO: split and run multiple times?
	}

	var deviceAction string
	var found bool
	if deviceAction, found = blePacketsTypes[action]; !found {
		deviceAction = "Unknown"
	}

	switch da := deviceAction; da {
	case "nearby":
		// nearby
		jApple = processNearby(data)

	case "handoff":
		// handoff
		jApple = processHandoff(data)

	//case "nearby_action":
	// nearby_action
	//jApple = processNearbyAction(data)

	//case "hey_siri":
	// hey_siri
	//jApple = processHeySiri(data)

	case "airpods":
		// airpods
		jApple = processAirpods(data)

	default:
		log.Printf("WARNING: No parser for action Apple: (0x%02x) %+v", action, data)

		type Apple struct {
			UnknownPacket `json:"unknown"`
		}

		pkt := Apple{
			UnknownPacket{
				Action: action,
				Len:    length,
				Data:   []byte(data), // this gets Base64 encoded by Marshall
			},
		}

		// convert to json
		var mpkt []byte
		var err error
		mpkt, err = json.Marshal(pkt)
		if err != nil {
			log.Println(err)
		}

		jApple = mpkt
	}

	ret := "{\"apple\":" + string(jApple) + "}"

	return ret
}
