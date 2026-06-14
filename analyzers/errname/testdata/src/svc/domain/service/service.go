// Package service — /domain/service is out of scope of GID-234
// (declaring an error here is the GID-144 area).
package service

import "errors"

var ErrNotFound = errors.New("not found")
