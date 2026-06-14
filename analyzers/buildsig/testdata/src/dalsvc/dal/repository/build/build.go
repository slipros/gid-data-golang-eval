// Eval GID-212: the contract of build functions in /dal/repository/build.
package build

import (
	"pgx/batch"

	"github.com/Masterminds/squirrel"
)

// --- Negative class: correct signatures pass ---

// Single query: (sql string, args []any, err error) — OK.
func SelectJobs(status string) (string, []any, error) {
	return "SELECT 1", []any{status}, nil
}

// Batch operation: (*batch.Batch, error) — OK.
func InsertJobsBatch(ids []string) (*batch.Batch, error) {
	return &batch.Batch{}, nil
}

// squirrel is imported and used in a build package — OK (check 2 does not apply here).
func buildSquirrel() (string, []any, error) {
	return squirrel.Select("id").ToSql()
}

// --- Positive class: a signature contract violation is caught ---

// Returns (string, error) — matches neither contract.
func BuildBad(status string) (string, error) { // want `GID-212: a build function must return \(sql string, args \[\]any, err error\) or \(\*batch\.Batch, error\)\. Fix: adjust the signature`
	return "", nil
}

// Returns *squirrel.SelectBuilder — a builder is not allowed as a result.
func BuildBuilder() *squirrel.SelectBuilder { // want `GID-212: a build function must return \(sql string, args \[\]any, err error\) or \(\*batch\.Batch, error\)\. Fix: adjust the signature`
	b := squirrel.Select("id")
	return &b
}

// --- Edge class ---

// A function without results — a violation (empty result list).
func BuildVoid() { // want `GID-212: a build function must return \(sql string, args \[\]any, err error\) or \(\*batch\.Batch, error\)\. Fix: adjust the signature`
}

// An unexported helper with a different signature — not flagged.
func helper(n int) int { return n }
