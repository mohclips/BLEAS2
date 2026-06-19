package services

import (
	"encoding/hex"
)

// SwiftPair is Microsoft's Windows BLE pairing-advertisement format
// (service UUID 0xfd5a). The service-data layout isn't fully documented
// publicly; we surface the bytes split into a header byte and the rest so
// patterns become visible across captures without claiming false structure.
type SwiftPair struct {
	Header int    `json:"header"`
	Data   string `json:"data"`
}

func parseSwiftPair(data []byte) []byte {
	sp := SwiftPair{}
	if len(data) >= 1 {
		sp.Header = int(data[0])
	}
	if len(data) >= 2 {
		sp.Data = hex.EncodeToString(data[1:])
	}
	return raw(sp)
}
