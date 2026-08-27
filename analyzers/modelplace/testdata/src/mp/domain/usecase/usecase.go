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

// --- Part C in /domain/usecase ---

// digest — an unexported data struct of the usecase layer.
type digest struct {
	total int
}

// summarize — positive: the struct leaves the package inside a map value.
func summarize(rows []string) map[string]digest { // want `GID-270: function "summarize" returns "digest" — a struct declared in this package, and /domain/usecase holds no data types of its own\. Fix: declare the returned type in /domain/model \(or //nolint:gidmodelplace when explicitly intended\)`
	return nil
}

// Refresh — negative: a value-returning constructor marks "Report" as the
// layer's entity, so handing it out is legal.
func Refresh() Report {
	return NewReport()
}
