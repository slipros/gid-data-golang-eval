package service

import "github.com/pkg/errors"

var ErrServiceLocal = errors.New("service local") // want `GID-169:`

type Snapshot struct{}
