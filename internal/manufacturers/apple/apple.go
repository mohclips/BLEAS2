package apple

import (
	"encoding/json"
	"fmt"
	"sync"

	log "github.com/mohclips/BLEAS2/internal/logging"
	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// reportedUnknown tracks subtype bytes we've already logged a warning for, so
// the log stays readable when an undocumented subtype is broadcast constantly.
var reportedUnknown sync.Map

// ParseMF parses an Apple (0x004c) manufacturer-data block. Apple Continuity
// allows multiple subtypes concatenated as [action, length, data...] runs;
// we walk them all and merge each parsed subtype into one `{"apple": {...}}`
// object keyed by subtype name. Unknown subtypes are surfaced under
// `unknown_0xNN` keys so different undocumented variants don't collide.
func ParseMF(mmfData []byte) string {
	if len(mmfData) < 4 {
		return ""
	}
	// Strip the leading 2-byte Manufacturer ID.
	mfData := mmfData[2:]

	combined := map[string]json.RawMessage{}
	i := 0
	for i+2 <= len(mfData) {
		action := int(mfData[i])
		length := int(mfData[i+1])
		if length == 0 || i+2+length > len(mfData) {
			if i+2 < len(mfData) {
				log.Warn("Apple truncated subtype: action=0x%02x claimed_length=%d remaining=%d", action, length, len(mfData)-i)
			}
			break
		}
		data := mfData[i+2 : i+2+length]

		body := dispatchAction(action, length, data)
		if body != nil {
			// Merge {"name": <inner>} into combined.
			var entry map[string]json.RawMessage
			if err := json.Unmarshal(body, &entry); err != nil {
				log.Error("apple: cannot remerge subtype 0x%02x output: %s", action, err)
			} else {
				for k, v := range entry {
					combined[k] = v
				}
			}
		}
		i += 2 + length
	}

	if len(combined) == 0 {
		return ""
	}
	merged, _ := json.Marshal(combined)
	return `{"apple":` + string(merged) + `}`
}

// dispatchAction parses one subtype block. Returns the wrapped JSON the inner
// parser produces (e.g. `{"ibeacon":{...}}`), or a synthetic
// `{"unknown_0xNN":{...}}` wrapper for undocumented subtypes.
func dispatchAction(action, length int, data []byte) []byte {
	switch blePacketsTypes[action] {
	case "airprint":
		return processAirPrint(data)
	case "airdrop":
		return processAirDrop(data)
	case "homekit":
		return processHomeKit(data)
	case "airpods":
		return processAirpods(data)
	case "hey_siri":
		return processHeySiri(data)
	case "airplay":
		return processAirPlayTarget(data)
	case "airplay_source":
		return processAirPlaySource(data)
	case "watch_c":
		return processMagicSwitch(data)
	case "handoff":
		return processHandoff(data)
	case "wifi_set":
		return processTetheringTarget(data)
	case "hotspot":
		return processTetheringSource(data)
	case "nearby_action":
		return processNearbyAction(data)
	case "nearby":
		return processNearby(data)
	case "iBeacon":
		return processiBeacon(data)
	case "findmy":
		return processFindMy(data)
	}

	// Unknown / undocumented subtype: emit the raw bytes under a key that
	// includes the hex code so distinct unknowns don't collide. Warn once
	// per subtype-byte so the operator notices new variants without the log
	// being swamped on repeats.
	if _, seen := reportedUnknown.LoadOrStore(action, true); !seen {
		log.Warn("No parser for action Apple: (0x%02x) %+v — bytes captured under unknown_0x%02x", action, data, action)
	}
	key := fmt.Sprintf("unknown_0x%02x", action)
	payload := mf.MarshalOrEmpty(UnknownPacket{
		Action: action,
		Len:    length,
		Data:   data,
	})
	return []byte(`{"` + key + `":` + string(payload) + `}`)
}
