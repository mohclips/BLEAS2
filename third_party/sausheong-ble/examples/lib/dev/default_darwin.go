package dev

import (
	"github.com/sausheong/ble"
	"github.com/sausheong/ble/darwin"
)

// DefaultDevice ...
func DefaultDevice(opts ...ble.Option) (d ble.Device, err error) {
	return darwin.NewDevice(opts...)
}
