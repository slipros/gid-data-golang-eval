// Package datalab — the module that owns the entities other modules depend on.
package datalab

// Module — a composed module handing out its services.
type Module struct {
	showcase *ShowcaseService
	saved    *SavedDatasetService
	filter   *FilterService
}

// NewLab builds the module.
func NewLab() *Module {
	return &Module{showcase: &ShowcaseService{}, saved: &SavedDatasetService{}, filter: &FilterService{}}
}

// ShowcaseService — one entity satisfying several consumer interfaces.
func (m *Module) ShowcaseService() *ShowcaseService { return m.showcase }

// SavedDatasetService — another entity.
func (m *Module) SavedDatasetService() *SavedDatasetService { return m.saved }

// FilterService — an entity of its own for the filtering role.
func (m *Module) FilterService() *FilterService { return m.filter }

// ShowcaseFor — a getter taking an argument: its value is not stable by inspection.
func (m *Module) ShowcaseFor(_ string) *ShowcaseService { return m.showcase }

// ShowcaseService implements both showcase-side interfaces of the consumer.
type ShowcaseService struct{}

// ShowcasesWithAttributeDataTypes — the first role.
func (s *ShowcaseService) ShowcasesWithAttributeDataTypes() []string { return nil }

// FilterShowcasesByAccess — the second role.
func (s *ShowcaseService) FilterShowcasesByAccess(_ string) []string { return nil }

// SavedDatasetService — the dataset entity.
type SavedDatasetService struct{}

// SavedDataset returns a dataset name.
func (s *SavedDatasetService) SavedDataset(_ string) string { return "" }

// FilterService — a separate entity for the filtering role.
type FilterService struct{}

// FilterShowcasesByAccess — the filtering role.
func (s *FilterService) FilterShowcasesByAccess(_ string) []string { return nil }
