// Eval GID-212: squirrel in /domain/service is banned (build packages only).
package service

import (
	"github.com/Masterminds/squirrel" // want `GID-212: squirrel is allowed only in repository build packages \(/dal/repository/build\)\. Fix: move squirrel usage into /dal/repository/build`
)

// --- Not-applicable class: the signature check does not apply outside build ---

// An arbitrary signature in a service — not flagged.
func Process(id string) (bool, error) { return true, nil }

func use() squirrel.SelectBuilder { return squirrel.Select("id") }
