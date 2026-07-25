// Boundary: database/sql imported only for its NULL value types — an entity
// converter is not storage access.
package convert

import "database/sql"

func TextOrEmpty(v sql.NullString) string {
	if !v.Valid {
		return ""
	}

	return v.String
}
