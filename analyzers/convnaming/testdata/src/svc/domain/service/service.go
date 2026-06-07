// Eval GID-135: конвертеры живут в convert/.
package service

import "context"

type Snapshot struct{ Name string }

type Row struct{ Name string }

// Позитив: функция-конвертер в самом service-пакете.
func ModelSnapshotFromRow(in *Row) Snapshot { // want `GID-135: converter "ModelSnapshotFromRow" must live in a convert/ subpackage of its layer`
	return Snapshot{Name: in.Name}
}

// Граничный кейс: ctx-helper — не конвертер (GID-166).
func SessionFromContext(ctx context.Context) (string, bool) {
	s, ok := ctx.Value(struct{}{}).(string)
	return s, ok
}

// Негатив: обычные функции паттерну не соответствуют.
func NewSnapshot() *Snapshot { return &Snapshot{} }
