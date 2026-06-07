// Eval для GID-195: settings.exclude — "Функция" и "Тип.Метод".
package service

import "exsvc/domain/model"

type Service struct{}

func legacyTitle(s *model.Snapshot) string { // исключена настройкой по имени
	return s.Name
}

func (s *Service) legacyRender(snap *model.Snapshot) string { // исключён как Тип.Метод
	return snap.Name
}

func otherTitle(s *model.Snapshot) string { // want `GID-195: приватная функция "otherTitle" работает только со значением model.Snapshot — это поведение модели: оформите её публичным методом этого типа`
	return s.ID
}
