// Eval для GID-195 (modelmethod): сервисный пакет — scope правила.
package service

import (
	"strings"

	"svc/domain/model"
)

type SnapshotService struct {
	prefix string
}

// --- Позитив: приватная функция с единственным model-параметром ---

func snapshotTitle(s *model.Snapshot) string { // want `GID-195: private function "snapshotTitle" works only with the model.Snapshot value\. Fix: this is model behaviour, make it a public method of that type`
	return strings.ToUpper(s.Name)
}

// --- Позитив: model-enum по значению ---

func isDone(st model.Status) bool { // want `GID-195: private function "isDone" works only with the model.Status value\. Fix: this is model behaviour, make it a public method of that type`
	return st == model.StatusDone
}

// --- Позитив: метод, не использующий ресивер ---

func (s *SnapshotService) renderSnapshot(snap *model.Snapshot) string { // want `GID-195: method "renderSnapshot" ignores its receiver and works only with the model.Snapshot value\. Fix: this is model behaviour, make it a public method of that type`
	return snap.ID + ":" + snap.Name
}

// --- Позитив: безымянный ресивер ---

func (*SnapshotService) pingSnapshot(s *model.Snapshot) bool { // want `GID-195: method "pingSnapshot" ignores its receiver and works only with the model.Snapshot value\. Fix: this is model behaviour, make it a public method of that type`
	return s.ID != ""
}

// --- Негатив: метод использует ресивер — легитимно принадлежит структуре ---

func (s *SnapshotService) decorate(snap *model.Snapshot) string {
	return s.prefix + snap.Name
}

// --- Негатив: два значения одного типа ---

func equalSnapshots(a, b *model.Snapshot) bool {
	return a.ID == b.ID
}

// --- Негатив: функция зависит от package-level символа своего пакета ---

const serviceTag = "svc:"

func tagSnapshot(s *model.Snapshot) string {
	return serviceTag + s.Name
}

// --- Негатив: результат — тип своего пакета, непереносима ---

func wrapSnapshot(s *model.Snapshot) *SnapshotService {
	return &SnapshotService{prefix: s.Name}
}

// --- Граница: variadic ---

func joinSnapshots(ss ...model.Snapshot) string {
	names := make([]string, 0, len(ss))
	for i := range ss {
		names = append(names, ss[i].Name)
	}
	return strings.Join(names, ",")
}

// --- Граница: слайс model-типа — не одно значение ---

func firstName(ss []model.Snapshot) string {
	if len(ss) == 0 {
		return ""
	}
	return ss[0].Name
}

// --- Граница: интерфейс model-слоя — метод не добавить ---

func validateAny(v model.Validator) bool {
	return v.Validate() == nil
}

// --- Граница: параметр типа своего пакета — не model ---

type ServiceOptions struct {
	Name string
}

func optionsName(o *ServiceOptions) string {
	return o.Name
}

// --- Граница: generic-функция ---

func anyTitle[T any](v T) T {
	return v
}

// --- Граница: экспортируемая функция — вне правила ---

func TitleSnapshot(s *model.Snapshot) string {
	return s.Name
}
