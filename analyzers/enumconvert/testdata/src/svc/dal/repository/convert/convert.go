// Неприменимость: convert-пакет без map-индексаций enum — диагностик нет.
package convert

type model struct{ Name string }

type entity struct{ Name string }

// Обычный полевой конвертер — никаких мап.
func EntityNameFromModel(in *model) entity {
	return entity{Name: in.Name}
}

func ModelNameFromEntity(in *entity) model {
	return model{Name: in.Name}
}
