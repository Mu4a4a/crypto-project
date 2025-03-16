package entities

import "github.com/pkg/errors"

var (
	ErrInvalidParam = errors.New("invalid param")
	ErrStorage      = errors.New("storage error")
	ErrProvider     = errors.New("provider error")
)
