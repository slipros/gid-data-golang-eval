// Eval для GID-194 (constscope): сервисный пакет — обычный scope правила.
package service

// --- Позитив: экспортируемая константа вне model/entity ---

const DefaultPageSize = 25 // want `GID-194: exported constant "DefaultPageSize" is declared outside model/entity\. Fix: keep shared constants in /domain/model or /dal/entity, and declare local ones where they are used`

// --- Позитив: константа используется только одним методом ---

const snapshotPrefix = "snap-" // want `GID-194: constant "snapshotPrefix" is used only in "Snapshot\.Render"\. Fix: declare it inside that function`

// --- Негатив: константа разделяется двумя методами — package-level легален ---

const snapshotTable = "snapshots"

type Snapshot struct{}

func (s *Snapshot) Render() string {
	return snapshotPrefix + snapshotTable
}

func (s *Snapshot) Table() string {
	return snapshotTable
}

// --- Негатив: константа объявлена внутри функции — целевое состояние ---

func (s *Snapshot) Endpoint() string {
	const endpoint = "/v1/snapshots"
	return endpoint
}
