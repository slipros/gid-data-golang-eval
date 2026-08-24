package app

import (
	"testing"

	"svc/marketplace"
)

// showcaseDouble stands in for every showcase interface of the constructor:
// one double satisfying them all is how a test is written (GID-250).
type showcaseDouble struct{}

func (showcaseDouble) ShowcasesWithAttributeDataTypes() []string { return nil }

func (showcaseDouble) FilterShowcasesByAccess(_ string) []string { return nil }

func (showcaseDouble) SavedDataset(_ string) string { return "" }

// TestModule wires the double into all three parameters — not judged.
func TestModule(t *testing.T) {
	double := showcaseDouble{}
	if marketplace.NewModule(double, double, double) == nil {
		t.Fatal("no module")
	}
}
