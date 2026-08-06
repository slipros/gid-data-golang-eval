// Package model holds the transaction type of the service (GID-175).
package model

import "context"

// InTransactionFunc — the transaction runner handed down from the composition
// root: the shape GID-260 matches by signature, whatever a project names it.
type InTransactionFunc func(ctx context.Context, fn func(ctx context.Context) error) error
