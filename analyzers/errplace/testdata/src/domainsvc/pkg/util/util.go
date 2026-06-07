// Неприменимость: пакет вне domain-дерева — правило не действует.
package util

import "errors"

var ErrUtil = errors.New("util")
