package gderr

import "errors"

func NewUnhandledValueError(value any) error { return errors.New("unhandled value") }
