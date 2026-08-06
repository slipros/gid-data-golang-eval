// Package service — a CORE service: the module takes core data through it.
package service

import "context"

// Integration — the core integration service.
type Integration struct{}

// Integration reads the core record.
func (s *Integration) Integration(ctx context.Context, id string) (string, error) { return id, nil }
