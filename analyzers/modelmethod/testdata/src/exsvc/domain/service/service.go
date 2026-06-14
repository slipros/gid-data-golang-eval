// Eval for GID-195: settings.exclude — "Function" and "Type.Method".
package service

import "exsvc/domain/model"

type Service struct{}

func legacyTitle(s *model.Snapshot) string { // excluded by name via settings
	return s.Name
}

func (s *Service) legacyRender(snap *model.Snapshot) string { // excluded as Type.Method
	return snap.Name
}

func otherTitle(s *model.Snapshot) string { // want `GID-195: private function "otherTitle" works only with the model.Snapshot value\. Fix: this is model behaviour, make it a public method of that type`
	return s.ID
}
