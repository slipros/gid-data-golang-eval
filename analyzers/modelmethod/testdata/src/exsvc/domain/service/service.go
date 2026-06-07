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

func otherTitle(s *model.Snapshot) string { // want `GID-195: private function "otherTitle" works only with the model.Snapshot value\. Fix: this is model behaviour, make it a public method of that type`
	return s.ID
}
