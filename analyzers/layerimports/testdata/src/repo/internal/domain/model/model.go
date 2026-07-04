// Shared entity in repo/internal/domain/model: importing it from
// repo/pkg/<module>/** is legal common-entity access (module.md), not a
// same-module import — the layer matrix does not apply to it.
package model

// Shared is a common entity available to every application module.
type Shared struct {
	ID string
}
