package errorx

import "errors"

var (
	ErrInvalidInput = errors.New("hypervisor: invalid input")
	ErrConflict     = errors.New("hypervisor: conflict")
	ErrNotFound     = errors.New("hypervisor: not found")
	ErrUnauthorized = errors.New("hypervisor: unauthorized")
	ErrUnavailable  = errors.New("hypervisor: unavailable")
)
