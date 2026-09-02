// Package repository — the data layer of the fixture service.
package repository

// DatasetSnapshot — the Iceberg read repository.
type DatasetSnapshot struct{}

// Snapshot returns the snapshot identifier.
func (d *DatasetSnapshot) Snapshot(id string) string { return id }

// Schema returns the schema of the table.
func (d *DatasetSnapshot) Schema(id string) string { return id }

// SnapshotManifest — the manifest read repository.
type SnapshotManifest struct{}

// Manifest returns the manifest body.
func (s *SnapshotManifest) Manifest(id string) string { return id }
