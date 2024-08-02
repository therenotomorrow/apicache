package domain

import "errors"

var (
	ErrKeyNotExist    = errors.New("key not exist")
	ErrKeyExpired     = errors.New("key is expired")
	ErrConnTimeout    = errors.New("connection timeout")
	ErrContextTimeout = errors.New("context timeout")
	ErrClosed         = errors.New("closed instance")
	ErrEmptyKey       = errors.New("empty key")
	ErrEmptyVal       = errors.New("empty value")
	ErrDataCorrupted  = errors.New("data corrupted")
)
