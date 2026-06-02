package repository

import (
	"errors"
	"time"
)

const (
	queryTimout = 5 * time.Second
)

var (
	ErrNotFound = errors.New("not found")
	ErrDuplicateEmail = errors.New("email already exists")
)