// Package services parses BLE Service Data advertisements (the per-service
// payloads exposed by a.ServiceData() in the BLE library).
//
// Coverage today:
//   - 0xfd6f Google Exposure Notification (COVID apps) — full RPI parse
//   - 0xfeaa Eddystone — UID / URL / TLM / EID subframes
//   - 0xfef3 Google Fast Pair (newer service UUID)
//   - 0xfe2c Google Fast Pair (older service UUID)
//   - 0xfd5a Microsoft Swift Pair
//
// Other observed UUIDs are surfaced via Name() with their canonical name
// when known (Battery, Heart Rate, Current Time, etc.) so the JSONL has a
// human-readable label even when no deep parser exists.
package services

import (
	"encoding/json"
	"strings"

	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
)

// Name returns the canonical name of a BLE service UUID (16-bit hex string
// in lower case, e.g. "180f"). Returns "" if unknown.
func Name(uuid string) string {
	if v, ok := serviceNames[strings.ToLower(uuid)]; ok {
		return v
	}
	return ""
}

// Parse dispatches by service UUID. data is the raw service-data bytes
// after the 16-bit UUID prefix has been stripped by the BLE library.
// Returns (name, json) — name is the structured key under which the parsed
// payload should appear; nil json means no deep parse, only the name (if
// known) should be surfaced.
func Parse(uuid string, data []byte) (string, []byte) {
	uuid = strings.ToLower(uuid)
	switch uuid {
	case "fd6f":
		return "exposure_notification", parseExposureNotification(data)
	case "feaa":
		return "eddystone", parseEddystone(data)
	case "fef3", "fe2c":
		return "fast_pair", parseFastPair(data, uuid)
	case "fd5a":
		return "swift_pair", parseSwiftPair(data)
	case "fe65", "fe33":
		return "chipolo", parseChipolo(data, uuid)
	}
	// Unknown service — emit raw bytes under a synthetic key so different
	// UUIDs don't collide.
	return "service_" + uuid, mf.MarshalOrEmpty(unknownServiceData{
		UUID: uuid,
		Name: Name(uuid),
		Data: data,
	})
}

type unknownServiceData struct {
	UUID string `json:"uuid"`
	Name string `json:"name,omitempty"`
	Data []byte `json:"data"`
}

// raw lets a parser return a json.RawMessage by wrapping anything that
// MarshalJSON's correctly. Kept as a sugar helper for the inner parsers.
func raw(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
