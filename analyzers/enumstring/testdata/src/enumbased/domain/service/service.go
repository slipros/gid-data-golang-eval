// Неприменимость: /domain/service вне scope GID-123 — диагностик нет.
package service

type ConsentEventType = string

type Status int

const (
	StatusA Status = 1
	StatusB Status = 2
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)
