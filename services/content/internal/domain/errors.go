package domain

import "errors"

var (
	// Content errors
	ErrContentNotFound     = errors.New("content not found")
	ErrContentExists       = errors.New("content with this slug already exists")
	ErrInvalidContentType  = errors.New("invalid content type")
	ErrInvalidContentInput = errors.New("invalid content input")

	// User errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserExists        = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")

	// Media errors
	ErrInvalidMediaType = errors.New("invalid media type")
	ErrMediaTooLarge    = errors.New("media file too large")
	ErrMediaUploadFailed = errors.New("media upload failed")

	// General errors
	ErrInternalServer = errors.New("internal server error")
	ErrInvalidInput   = errors.New("invalid input")
	ErrDatabase       = errors.New("database error")
)
