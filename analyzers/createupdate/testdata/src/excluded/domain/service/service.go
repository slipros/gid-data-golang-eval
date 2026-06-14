// Eval for settings.exclude: the listed methods are not reported.
package service

import "context"

type Session struct{ ID string }

type Job struct{}

// Excluded as "Job.CreateJob" (Type.Method).
func (j *Job) CreateJob(ctx context.Context, name string) (Session, error) {
	return Session{}, nil
}

// Excluded as "UpdateSession" (a method name).
func (j *Job) UpdateSession(ctx context.Context, id string) (Session, error) {
	return Session{}, nil
}

// Not excluded — reported.
func (j *Job) CreateSession(ctx context.Context) (Session, error) { // want `GID-112: method "CreateSession" creates/updates state and must return only error`
	return Session{}, nil
}
