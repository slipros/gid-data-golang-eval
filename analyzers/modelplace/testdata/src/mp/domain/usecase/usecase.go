// Eval for GID-270 (model-place), part B: the same rule in /domain/usecase.
package usecase

// AltCraftTriggerBuild — positive: an exported data struct declared right in
// usecase, among the entities.
type AltCraftTriggerBuild struct { // want `GID-270: data struct "AltCraftTriggerBuild" is declared in /domain/usecase — it has no methods and is built by no constructor, so it is a data model\. Fix: move the type to /domain/model \(or //nolint:gidmodelplace when explicitly intended\)`
	TemplateID string
	Channels   []string
	Debounced  bool
}

// Report — boundary: a constructor returning a VALUE (not a pointer) also
// marks the type as the layer's entity.
type Report struct {
	lines []string
}

// NewReport builds the report.
func NewReport() Report {
	return Report{}
}
