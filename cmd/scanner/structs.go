package main

// ParsedManufacturerData - The parsed data we are after
type ParsedManufacturerData struct {
	ID   uint16 `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// we add more here once its Marshalled as json
	Details string `json:"details,omitempty"`
}

// Device - represents a BLE device, with our parsed data tacked on
type Device struct {
	Timestamp        string                 `json:"@timestamp"`
	Address          string                 `json:"address"`
	Detected         string                 `json:"detected"`
	Since            string                 `json:"since,omitempty"`
	Name             string                 `json:"name,omitempty"`
	RSSI             int                    `json:"rssi"`
	Advertisement    string                 `json:"advertisement,omitempty"`
	ScanResponse     string                 `json:"scanresponse,omitempty"`
	ManufacturerData ParsedManufacturerData `json:"manufacturerdata,omitempty"`
}
