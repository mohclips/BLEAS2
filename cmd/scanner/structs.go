package main

import "encoding/json"

// ParsedManufacturerData - The parsed data we are after
type ParsedManufacturerData struct {
	ID      uint16          `json:"id,omitempty"`
	Name    string          `json:"name,omitempty"`
	Details json.RawMessage `json:"details,omitempty"`
}

// ParsedServiceData - Service data if present
type ParsedServiceData struct {
	ID      string          `json:"id,omitempty"`
	Name    string          `json:"name,omitempty"`
	UUIDs   []string        `json:"uuids,omitempty"`
	Details json.RawMessage `json:"details,omitempty"`
}

// CommonData carries the identity of the observed advertiser. Per-packet
// fields (rssi, detected) have moved into Observation; what remains here is
// the stable per-window identity plus the window boundaries.
type CommonData struct {
	Address       string `json:"address"`
	AddressType   string `json:"address_type"`
	FirstSeen     string `json:"first_seen"`
	LastSeen      string `json:"last_seen"`
	Name          string `json:"name,omitempty"`
	Advertisement string `json:"advertisement,omitempty"`
	ScanResponse  string `json:"scanresponse,omitempty"`
}

// Observation summarises every RSSI sample collected during a dedup
// window. Even when only one sample arrived, count=1 and the min/max/mean
// match the single sample — that keeps the schema regular for the
// downstream reporter regardless of whether dedup batching is on.
type Observation struct {
	Count       int     `json:"count"`
	RSSIMin     int     `json:"rssi_min"`
	RSSIMax     int     `json:"rssi_max"`
	RSSIMean    float64 `json:"rssi_mean"`
	RSSISamples []int   `json:"rssi_samples"`
}

// Device represents a BLE advertiser as seen during one observation window.
type Device struct {
	Timestamp        string                 `json:"@timestamp"` // = Common.FirstSeen, kept for ES-style consumers
	Common           CommonData             `json:"Common"`
	Observation      Observation            `json:"observation"`
	ManufacturerData ParsedManufacturerData `json:"manufacturerdata,omitempty"`
	ServiceData      ParsedServiceData      `json:"servicedata,omitempty"`
}

// MACaddressTypes - address type
var MACaddressTypes = map[int]string{
	0x00: "public", // - Use the controller's public address.
	0x01: "random", // - Use a generated static address.
	0x02: "RPA",    // - Use resolvable private addresses.
	0x03: "NRPA",   // - Use non-resolvable private addresses.
}

// EventTypes - advertisement event types
var EventTypes = map[int]string{
	0x00: "ADV_IND",         // - connectable and scannable undirected advertising
	0x01: "ADV_DIRECT_IND",  // - connectable directed advertising
	0x02: "ADV_SCAN_IND",    // - scannable undirected advertising
	0x03: "ADV_NONCONN_IND", // - non-connectable undirected advertising
	0x04: "SCAN_RSP",        // - scan response
}
