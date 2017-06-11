package hidraw

import (
	"errors"
	"io"
)

// ErrNoDevice is returned if there is no device found for the given vendor and product ID.
var ErrNoDevice = errors.New("no device found")

// OpenInputDevice
func OpenInputDevice(vendor, product int) (io.ReadCloser, error) {
	return findAndOpen(vendor, product)
}
