// Граничный: та же конструкция вне convert-пакета — не матчится.
// Пакет service (последний сегмент пути не convert) → вне scope.
package service

type (
	EntityStatus string
	ModelStatus  string
)

var statusMap = map[EntityStatus]ModelStatus{"active": "active"}

// Та же enum-индексация мапы без comma-ok — но вне convert-пакета, не матчим.
func mapStatus(s EntityStatus) ModelStatus {
	return statusMap[s]
}
