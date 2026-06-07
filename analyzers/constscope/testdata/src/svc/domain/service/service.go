// Eval для GID-194 (constscope): сервисный пакет — обычный scope правила.
package service

// --- Позитив: экспортируемая константа вне model/entity ---

const DefaultPageSize = 25 // want `GID-194: экспортируемая константа "DefaultPageSize" объявлена вне model/entity — общие константы живут в /domain/model или /dal/entity, локальные объявляются там, где используются`

// --- Позитив: константа используется только одним методом ---

const snapshotPrefix = "snap-" // want `GID-194: константа "snapshotPrefix" используется только в "Snapshot.Render" — объявите её внутри этой функции`

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
