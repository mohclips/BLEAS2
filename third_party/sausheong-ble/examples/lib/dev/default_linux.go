package dev

import (
	"github.com/sausheong/ble"
	"github.com/sausheong/ble/linux"
)

// DefaultDevice ...
func DefaultDevice(opts ...ble.Option) (d ble.Device, err error) {
	return linux.NewDevice(opts...)
}
