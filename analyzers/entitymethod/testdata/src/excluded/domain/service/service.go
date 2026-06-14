// Eval for settings.exclude: the listed methods are not reported.
package service

import "context"

type Job struct{}

// Excluded as "Job.Close" (Type.Method) — otherwise it would be caught by check 3.
func (j *Job) Close() error {
	return nil
}

// Excluded as "Ping" (a method name).
func (j *Job) Ping(ctx context.Context) error {
	return nil
}

// Not excluded — caught by check 3 (no Job entity name).
func (j *Job) Flush(ctx context.Context) error { // want `GID-114: method name "Flush" must contain the entity name "Job"`
	return nil
}
