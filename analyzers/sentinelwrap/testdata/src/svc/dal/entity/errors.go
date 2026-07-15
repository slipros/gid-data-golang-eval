// dal errors (static) — their home is /dal/entity.
package entity

import "github.com/pkg/errors"

var ErrNoResult = errors.New("no result")
