// Неприменимость GID-197: model-слой вне scope — интерфейсы model могут
// описывать контракт для внешних потребителей.
package model

type Filterable interface {
	Apply(query string) string
	Reset()
}
