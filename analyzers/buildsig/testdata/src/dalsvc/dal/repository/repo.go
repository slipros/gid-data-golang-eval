// Eval GID-212: squirrel outside a build package (/dal/repository) is banned;
// arbitrary function signatures outside build are not flagged (not applicable).
package repository

import (
	"github.com/Masterminds/squirrel" // want `GID-212: squirrel is allowed only in repository build packages \(/dal/repository/build\)\. Fix: move squirrel usage into /dal/repository/build`
)

// --- Not-applicable class: the signature check does not apply outside build ---

// An exported function with an arbitrary signature in /dal/repository — not flagged.
func DoStuff(id string) (int, error) { return 0, nil }

// A function without results outside build — not flagged.
func Reset() {}

// squirrel is imported above — caught as an import violation.
func use() squirrel.SelectBuilder { return squirrel.Select("id") }
