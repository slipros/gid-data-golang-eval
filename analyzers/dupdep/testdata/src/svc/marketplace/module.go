// Package marketplace — the consumer declaring its dependency interfaces (GID-134).
package marketplace

// SavedDatasetService — the dataset dependency.
type SavedDatasetService interface {
	SavedDataset(id string) string
}

// ShowcasesWithAttributeDataTypesService — the showcase-listing dependency.
type ShowcasesWithAttributeDataTypesService interface {
	ShowcasesWithAttributeDataTypes() []string
}

// FilterShowcasesByAccessService — the showcase-filtering dependency.
type FilterShowcasesByAccessService interface {
	FilterShowcasesByAccess(user string) []string
}

// Module — the marketplace module.
type Module struct {
	saved     SavedDatasetService
	showcases ShowcasesWithAttributeDataTypesService
	filter    FilterShowcasesByAccessService
}

// NewModule builds the marketplace module.
func NewModule(
	saved SavedDatasetService,
	showcases ShowcasesWithAttributeDataTypesService,
	filter FilterShowcasesByAccessService,
) *Module {
	return &Module{saved: saved, showcases: showcases, filter: filter}
}

// Register wires a listener into the module — not a constructor.
func Register(showcases ShowcasesWithAttributeDataTypesService, filter FilterShowcasesByAccessService) {
	_, _ = showcases, filter
}
