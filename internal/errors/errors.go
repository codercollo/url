package errors

import "errors"

//URL lifecyle errors
var (
	ErrURLNotFound = errors.New("short code not found")
	ErrURLExpired  = errors.New("short code has expired")
	ErrURLInactive = errors.New("short code is inactive")
)

//Input validation errors
var (
	ErrEmptyurl       = errors.New("original URL cannot be empty")
	ErrInvalidURL     = errors.New("invalid URL format")
	ErrCodeTaken      = errors.New("custom short code already in use")
	ErrEmptyShortCode = errors.New("short code cannot be empty")
)
