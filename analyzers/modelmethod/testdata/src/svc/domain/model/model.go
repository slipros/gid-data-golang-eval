// Model-пакет для eval GID-195.
package model

type Snapshot struct {
	ID   string
	Name string
}

type Status string

const (
	StatusActive Status = "active"
	StatusDone   Status = "done"
)

type Validator interface {
	Validate() error
}
