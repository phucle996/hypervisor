package errorx

import "errors"

var (
	ErrInvalidResource = errors.New("hypervisor: invalid resource")
	ErrConflict        = errors.New("hypervisor: conflict")
	ErrNotFound        = errors.New("hypervisor: not found")
	ErrUnavailable     = errors.New("hypervisor: unavailable")
)
