// Eval for GID-197 (ifacemin): a service package is the rule's scope.
package service

import (
	"context"
	"io"
)

// --- Positive: an unused method of a dependency interface ---

type SnapshotRepository interface {
	Snapshot(ctx context.Context, id string) (string, error)
	CreateSnapshot(ctx context.Context, name string) error
	DeleteSnapshot(ctx context.Context, id string) error // want `GID-197: method "DeleteSnapshot" of interface "SnapshotRepository" is not used in the consumer package\. Fix: keep the interface minimal, remove the method`
}

type SnapshotService struct {
	repo SnapshotRepository
}

func NewSnapshotService(repo SnapshotRepository) *SnapshotService {
	return &SnapshotService{repo: repo}
}

func (s *SnapshotService) Render(ctx context.Context, id string) (string, error) {
	return s.repo.Snapshot(ctx, id)
}

func (s *SnapshotService) Create(ctx context.Context, name string) error {
	return s.repo.CreateSnapshot(ctx, name)
}

// --- Negative: a method value is a use too ---

type SnapshotCache interface {
	Get(id string) (string, bool)
	Warm()
}

type cacheRunner struct {
	cache SnapshotCache
}

func (c *cacheRunner) run() string {
	warm := c.cache.Warm
	warm()
	v, _ := c.cache.Get("1")
	return v
}

// --- Boundary: the value escapes into any — the interface is skipped entirely ---

type AuditSink interface {
	Write(msg string)
	Flush() error // not called, but the sink leaves under a different type — skipped
}

func (s *SnapshotService) Audit(sink AuditSink) {
	sink.Write("audit")
	var bucket any = sink
	_ = bucket
}

// --- Boundary: an embedded interface of the same package — use through
// the outer interface counts ---

type snapshotReader interface {
	ReadSnapshot() string
}

type snapshotReadWriter interface {
	snapshotReader
	WriteSnapshot(v string)
}

func consumeRW(rw snapshotReadWriter) string {
	rw.WriteSnapshot("x")
	return rw.ReadSnapshot()
}

// --- Boundary: an embedded standard library interface is not checked ---

type SnapshotSource interface {
	io.Closer
	Open(name string) error
}

func useSource(src SnapshotSource) error {
	if err := src.Open("x"); err != nil {
		return err
	}
	return src.Close()
}

// --- Boundary: use only from *_test.go — a violation ---

type SnapshotProbe interface {
	Ping() error // want `GID-197: method "Ping" of interface "SnapshotProbe" is not used in the consumer package\. Fix: keep the interface minimal, remove the method`
}

type prober struct {
	probe SnapshotProbe
}
