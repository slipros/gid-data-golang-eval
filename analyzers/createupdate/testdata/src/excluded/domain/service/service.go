// Eval для settings.exclude: перечисленные методы не репортятся.
package service

import "context"

type Session struct{ ID string }

type Job struct{}

// Исключён как "Job.CreateJob" (Тип.Метод).
func (j *Job) CreateJob(ctx context.Context, name string) (Session, error) {
	return Session{}, nil
}

// Исключён как "UpdateSession" (имя метода).
func (j *Job) UpdateSession(ctx context.Context, id string) (Session, error) {
	return Session{}, nil
}

// Не исключён — репортится.
func (j *Job) CreateSession(ctx context.Context) (Session, error) { // want `GID-112: method "CreateSession" creates/updates state and must return only error`
	return Session{}, nil
}
