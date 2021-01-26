package main

// ParsedManufacturerData - The parsed data we are after
type ParsedManufacturerData struct {
	ID   uint16 `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// we add more here once its Marshalled as json
	Details string `json:"details,omitempty"`
}

// ParsedServiceData - Service data if present
type ParsedServiceData struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// we add more here once its Marshalled as json
	Details string `json:"details,omitempty"`
}

// CommonData - to all advertisements
type CommonData struct {
	Address       string `json:"address"`
	AddressType   string `json:"address_type"`
	Detected      string `json:"detected"`
	Since         string `json:"since,omitempty"`
	Name          string `json:"name,omitempty"`
	RSSI          int    `json:"rssi"`
	Advertisement string `json:"advertisement,omitempty"`
	ScanResponse  string `json:"scanresponse,omitempty"`
}

// Device - represents a BLE device, with our parsed data tacked on
type Device struct {
	Timestamp        string `json:"@timestamp"`
	Common           CommonData
	ManufacturerData ParsedManufacturerData `json:"manufacturerdata,omitempty"`
	ServiceData      ParsedServiceData      `json:"servicedata,omitempty"`
}

// MACaddressTypes - address type
var MACaddressTypes = map[int]string{
	0x00: "public", // - Use the controller’s public address.
	0x01: "random", // - Use a generated static address.
	0x02: "RPA",    // - Use resolvable private addresses.
	0x03: "NRPA",   // - Use non-resolvable private addresses.
}

// EventTypes - adverstiment event types
var EventTypes = map[int]string{
	0x00: "ADV_IND",         // - connectable and scannable undirected advertising
	0x01: "ADV_DIRECT_IND",  // - connectable directed advertising
	0x02: "ADV_SCAN_IND",    // - scannable undirected advertising
	0x03: "ADV_NONCONN_IND", // - non-connectable undirected advertising
	0x04: "SCAN_RSP",        // - scan response
}
