// Negative: error.go is the (sole, default) allowed file — declarations here are fine.
package entity

import "github.com/pkg/errors"

var ErrRowExpired = errors.New("row expired")

var ErrRowArchived = errors.New("row archived")
