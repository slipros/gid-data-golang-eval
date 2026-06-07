package svc

import "github.com/pkg/errors"

type Repo struct{}

// --- Позитив: errors.New в теле метода ---

func (r Repo) Find() error {
	return errors.New("not found in repo") // want `GID-136: errors.New at runtime`
}
