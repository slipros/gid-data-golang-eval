// Package usecase — the public /latest page and the ports it depends on.
package usecase

// LatestPageSnapshot — the snapshot resolver as the page sees it.
type LatestPageSnapshot interface {
	Snapshot(id string) string
}

// LatestPageSchema — the schema service as the page sees it.
type LatestPageSchema interface {
	Schema(id string) string
}

// LatestPageStore — the object store of the page.
type LatestPageStore interface {
	Object(key string) []byte
}

// LatestPageParquet — the parquet reader of the page.
type LatestPageParquet interface {
	Rows(data []byte) int
}

// LatestPageCatalog — the catalog of the page.
type LatestPageCatalog interface {
	Catalog() string
}

// LatestPage — the usecase of the public /latest page.
type LatestPage struct {
	snapshot LatestPageSnapshot
	schema   LatestPageSchema
	store    LatestPageStore
	parquet  LatestPageParquet
}

// NewLatestPage builds the usecase of the public /latest page.
func NewLatestPage(
	snapshot LatestPageSnapshot,
	schema LatestPageSchema,
	store LatestPageStore,
	parquet LatestPageParquet,
) *LatestPage {
	return &LatestPage{snapshot: snapshot, schema: schema, store: store, parquet: parquet}
}

// UseCatalog takes the catalog of the page.
func UseCatalog(catalog LatestPageCatalog) string { return catalog.Catalog() }
