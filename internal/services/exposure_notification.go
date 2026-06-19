package services

import (
	"encoding/hex"
)

// ExposureNotification: Google/Apple's GAEN protocol used by COVID contact-
// tracing apps (NHS COVID-19, Corona-Warn-App, etc.).
//
// Service data (under UUID 0xfd6f) is:
//
//   - bytes 0-15: Rolling Proximity Identifier (RPI) — rotates with the
//                 Temporary Exposure Key, every ~10 minutes
//   - bytes 16-19: Associated Encrypted Metadata (AEM) — TX power + version
//
// The RPI is privacy-preserving (rotates frequently and is deterministically
// derived from a per-day Temporary Exposure Key) but its mere presence
// confirms the device is running a GAEN-compatible app.
type ExposureNotification struct {
	RPI      string `json:"rpi"`
	Metadata string `json:"metadata"`
}

func parseExposureNotification(data []byte) []byte {
	en := ExposureNotification{}
	if len(data) >= 16 {
		en.RPI = hex.EncodeToString(data[0:16])
	}
	if len(data) >= 20 {
		en.Metadata = hex.EncodeToString(data[16:20])
	}
	return raw(en)
}
