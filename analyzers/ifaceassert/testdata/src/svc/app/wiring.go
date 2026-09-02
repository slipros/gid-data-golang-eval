// Package app — the composition root of the fixture service.
package app

import (
	"svc/repository"
	"svc/service"
	"svc/usecase"
)

// Compile-time contract check: the Iceberg read repository implements the
// resolver port — and the constructor below is handed the very same pointer.
var _ service.DatasetSnapshotRepository = (*repository.DatasetSnapshot)(nil) // want `GID\-274:\ redundant\ compile\-time\ assertion:\ the\ package\ already\ passes\ this\ value\ as\ service\.DatasetSnapshotRepository\ at\ wiring\.go:52,\ so\ the\ compiler\ checks\ the\ contract\ there\.\ Fix:\ delete\ the\ "var\ _\ service\.DatasetSnapshotRepository\ =\ \(\*repository\.DatasetSnapshot\)\(nil\)"\ line`

// The manifest port: the same repository fills a field of a composite literal.
var _ service.SnapshotManifestRepository = (*repository.SnapshotManifest)(nil) // want `GID\-274:\ redundant\ compile\-time\ assertion:\ the\ package\ already\ passes\ this\ value\ as\ service\.SnapshotManifestRepository\ at\ wiring\.go:54,\ so\ the\ compiler\ checks\ the\ contract\ there\.\ Fix:\ delete\ the\ "var\ _\ service\.SnapshotManifestRepository\ =\ \(\*repository\.SnapshotManifest\)\(nil\)"\ line`

// The schema port: the conversion happens in a plain assignment.
var _ service.TableSchemaRepository = (*repository.DatasetSnapshot)(nil) // want `GID\-274:\ redundant\ compile\-time\ assertion:\ the\ package\ already\ passes\ this\ value\ as\ service\.TableSchemaRepository\ at\ wiring\.go:59,\ so\ the\ compiler\ checks\ the\ contract\ there\.\ Fix:\ delete\ the\ "var\ _\ service\.TableSchemaRepository\ =\ \(\*repository\.DatasetSnapshot\)\(nil\)"\ line`

// Grouped assertions of the page, the shape the composition root is written in.
var (
	// Asserted as a value, wired as a value.
	_ usecase.LatestPageStore = latestPageStore{} // want `GID\-274:\ redundant\ compile\-time\ assertion:\ the\ package\ already\ passes\ this\ value\ as\ usecase\.LatestPageStore\ at\ wiring\.go:64,\ so\ the\ compiler\ checks\ the\ contract\ there\.\ Fix:\ delete\ the\ "var\ _\ usecase\.LatestPageStore\ =\ latestPageStore\{\}"\ line`
	// Asserted for the pointer, wired as a value: the method set of the pointer
	// holds the method set of the value, so the conversion proves the assertion.
	_ usecase.LatestPageParquet = (*latestPageParquet)(nil) // want `GID\-274:\ redundant\ compile\-time\ assertion:\ the\ package\ already\ passes\ this\ value\ as\ usecase\.LatestPageParquet\ at\ wiring\.go:64,\ so\ the\ compiler\ checks\ the\ contract\ there\.\ Fix:\ delete\ the\ "var\ _\ usecase\.LatestPageParquet\ =\ \(\*latestPageParquet\)\(nil\)"\ line`
	// Asserted as a value, wired as a pointer: the pointer's method set is the
	// wider one, so the conversion below says nothing about the value type.
	_ usecase.LatestPageCatalog = catalogValue{}
	// Nothing in this package hands the catalog service over: the assertion is
	// the only check there is (only the test file below converts it).
	_ usecase.LatestPageCatalog = (*service.DatasetCatalog)(nil)
)

type latestPageStore struct{}

func (latestPageStore) Object(key string) []byte { return []byte(key) }

type latestPageParquet struct{}

func (latestPageParquet) Rows(data []byte) int { return len(data) }

type catalogValue struct{}

func (catalogValue) Catalog() string { return "value" }

// Wire builds the public /latest page.
func Wire() *usecase.LatestPage {
	snapshotRepo := &repository.DatasetSnapshot{}
	manifestRepo := &repository.SnapshotManifest{}

	snapshot := service.NewDatasetSnapshot(snapshotRepo)

	registry := service.Registry{Manifests: manifestRepo}
	_ = registry

	var schemaRepo service.TableSchemaRepository

	schemaRepo = snapshotRepo
	schema := service.NewTableSchema(schemaRepo)

	_ = usecase.UseCatalog(&catalogValue{})

	return usecase.NewLatestPage(snapshot, schema, latestPageStore{}, latestPageParquet{})
}

// A named variable is not an assertion: this is the wiring itself, declared
// with an explicit interface type, and nothing repeats it.
var pageStore usecase.LatestPageStore = latestPageStore{}
