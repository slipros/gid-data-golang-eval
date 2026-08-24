// Package app — the composition root wiring modules together.
package app

import (
	"svc/datalab"
	"svc/marketplace"
)

// --- Class 1: positive ---

// BuildInline passes one entity under two interfaces, straight from the getter.
func BuildInline(lab *datalab.Module) *marketplace.Module {
	return marketplace.NewModule(
		lab.SavedDatasetService(),
		lab.ShowcaseService(),
		lab.ShowcaseService(), // want `GID-268: constructor NewModule receives the same value in parameters #2 showcases and #3 filter — one dependency passed twice under different interfaces\. Fix: merge ShowcasesWithAttributeDataTypesService and FilterShowcasesByAccessService into a single interface and take the dependency as one parameter`
	)
}

// BuildVar passes the same entity through a variable.
func BuildVar(lab *datalab.Module) *marketplace.Module {
	showcases := lab.ShowcaseService()

	return marketplace.NewModule(lab.SavedDatasetService(), showcases, showcases) // want `GID-268: constructor NewModule receives the same value in parameters #2 showcases and #3 filter — one dependency passed twice under different interfaces\. Fix: merge ShowcasesWithAttributeDataTypesService and FilterShowcasesByAccessService into a single interface and take the dependency as one parameter`
}

// deps — a wiring struct holding the entities.
type deps struct {
	showcase *datalab.ShowcaseService
	saved    *datalab.SavedDatasetService
}

// build passes the same field under two interfaces.
func (d *deps) build() *marketplace.Module {
	return marketplace.NewModule(d.saved, d.showcase, d.showcase) // want `GID-268: constructor NewModule receives the same value in parameters #2 showcases and #3 filter — one dependency passed twice under different interfaces\. Fix: merge ShowcasesWithAttributeDataTypesService and FilterShowcasesByAccessService into a single interface and take the dependency as one parameter`
}

// --- Class 2: negative ---

// BuildDistinct gives every interface its own entity.
func BuildDistinct(lab *datalab.Module) *marketplace.Module {
	return marketplace.NewModule(lab.SavedDatasetService(), lab.ShowcaseService(), lab.FilterService())
}

// BuildDistinctFields takes the entities from different wiring structs.
func BuildDistinctFields(left, right *deps) *marketplace.Module {
	return marketplace.NewModule(left.saved, left.showcase, right.showcase)
}

// --- Class 3: boundary ---

// BuildFromCallWithArgs repeats a call that takes an argument: nothing says the
// two calls return one and the same entity.
func BuildFromCallWithArgs(lab *datalab.Module) *marketplace.Module {
	return marketplace.NewModule(lab.SavedDatasetService(), lab.ShowcaseFor("a"), lab.ShowcaseFor("a"))
}

// BuildNil leaves the dependencies unset: nil is not a dependency passed twice.
func BuildNil() *marketplace.Module {
	return marketplace.NewModule(nil, nil, nil)
}

// --- Class 4: non-applicability ---

// RegisterTwice passes one entity into a function that is not a constructor.
func RegisterTwice(lab *datalab.Module) {
	marketplace.Register(lab.ShowcaseService(), lab.ShowcaseService())
}
