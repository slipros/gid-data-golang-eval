// Неприменимость: /domain/service — НЕ корень слоя model/entity,
// GID-169 здесь не действует (это зона GID-144). error-var в любом
// файле сервиса не должна давать диагностику GID-169.
package service

import "github.com/pkg/errors"

var ErrServiceLocal = errors.New("service local")

type Snapshot struct{}
