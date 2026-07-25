// Boundary: the whole dal layer is allowed, not just repository — an entity
// carries driver column types.
package entity

import "github.com/jackc/pgx/v5/pgtype"

type Member struct {
	Name pgtype.Text
}
