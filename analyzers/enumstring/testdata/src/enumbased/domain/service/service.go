// Not applicable: /domain/service is outside the GID-123 scope — no diagnostics.
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
