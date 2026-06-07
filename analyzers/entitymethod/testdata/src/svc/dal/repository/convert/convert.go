// Неприменимость: подпакет convert/ вне scope (scope — корень слоя).
package convert

import "context"

type Snapshot struct{ ID string }

type Mapper struct{}

// Метод-глагол без сущности и с префиксом List — но пакет вне scope, диагностики нет.
func (m *Mapper) ListSnapshots(ctx context.Context) ([]Snapshot, error) {
	return nil, nil
}
