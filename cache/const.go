package cache

import "errors"

var (
	// ErrNotFound error occurred in .Get() when key was not found
	ErrNotFound = errors.New("Key not found")
)
