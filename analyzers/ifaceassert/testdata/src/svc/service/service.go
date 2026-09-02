// Package service — the domain services of the fixture and the ports they
// reach the data layer through.
package service

// DatasetSnapshotRepository — the port the snapshot resolver goes through.
type DatasetSnapshotRepository interface {
	Snapshot(id string) string
}

// TableSchemaRepository — the port the schema service goes through.
type TableSchemaRepository interface {
	Schema(id string) string
}

// SnapshotManifestRepository — the port the manifest service goes through.
type SnapshotManifestRepository interface {
	Manifest(id string) string
}

// DatasetSnapshot — the snapshot resolver.
type DatasetSnapshot struct {
	repo DatasetSnapshotRepository
}

// NewDatasetSnapshot builds the snapshot resolver.
func NewDatasetSnapshot(repo DatasetSnapshotRepository) *DatasetSnapshot {
	return &DatasetSnapshot{repo: repo}
}

// Snapshot resolves the snapshot of the dataset.
func (d *DatasetSnapshot) Snapshot(id string) string { return d.repo.Snapshot(id) }

// TableSchema — the schema service.
type TableSchema struct {
	repo TableSchemaRepository
}

// NewTableSchema builds the schema service.
func NewTableSchema(repo TableSchemaRepository) *TableSchema {
	return &TableSchema{repo: repo}
}

// Schema returns the schema of the table.
func (t *TableSchema) Schema(id string) string { return t.repo.Schema(id) }

// DatasetCatalog — the catalog service.
type DatasetCatalog struct{}

// Catalog returns the name of the catalog.
func (d *DatasetCatalog) Catalog() string { return "iceberg" }

// Registry — the manifest ports of the service, filled by the composition root.
type Registry struct {
	Manifests SnapshotManifestRepository
}
