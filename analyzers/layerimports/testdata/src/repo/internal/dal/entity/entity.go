// A repo/internal/dal entity: a different module from repo/pkg/<module>'s
// point of view, so importing it from pkg/<module>/server is not banned by
// GID-224 even though its path contains the "dal" segment.
package entity

// Invoice is a plain dal entity.
type Invoice struct {
	ID string
}
