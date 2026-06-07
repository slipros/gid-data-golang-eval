// Eval для GID-197 (ifacemin): сервисный пакет — scope правила.
package service

import (
	"context"
	"io"
)

// --- Позитив: неиспользуемый метод интерфейса-зависимости ---

type SnapshotRepository interface {
	Snapshot(ctx context.Context, id string) (string, error)
	CreateSnapshot(ctx context.Context, name string) error
	DeleteSnapshot(ctx context.Context, id string) error // want `GID-197: метод "DeleteSnapshot" интерфейса "SnapshotRepository" не используется в пакете-потребителе — интерфейс минимален: уберите метод из интерфейса`
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

// --- Негатив: метод-значение — тоже использование ---

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

// --- Граница: значение уходит в any — интерфейс пропускается целиком ---

type AuditSink interface {
	Write(msg string)
	Flush() error // не вызывается, но sink уезжает под другим типом — пропуск
}

func (s *SnapshotService) Audit(sink AuditSink) {
	sink.Write("audit")
	var bucket any = sink
	_ = bucket
}

// --- Граница: embedded-интерфейс того же пакета — использование через
// внешний интерфейс засчитывается ---

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

// --- Граница: embedded-интерфейс стандартной библиотеки не проверяется ---

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

// --- Граница: использование только из *_test.go — нарушение ---

type SnapshotProbe interface {
	Ping() error // want `GID-197: метод "Ping" интерфейса "SnapshotProbe" не используется в пакете-потребителе — интерфейс минимален: уберите метод из интерфейса`
}

type prober struct {
	probe SnapshotProbe
}
