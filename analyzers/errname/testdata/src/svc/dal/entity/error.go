// Package entity — /dal/entity is out of scope of GID-234: generic names are the dal convention.
package entity

import "errors"

var (
	ErrNoResult      = errors.New("no result")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)
