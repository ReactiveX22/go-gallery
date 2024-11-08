package models

import "errors"

var (
	ErrEmailTaken = errors.New("models: email address is already taken")
	ErrNotFound   = errors.New("models: no resource could be found with the provied info")
)
