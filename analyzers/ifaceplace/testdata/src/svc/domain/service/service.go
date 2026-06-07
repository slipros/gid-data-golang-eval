// Eval для GID-134 (interface-near-consumer). Потребитель — слой
// /domain/service.
package service

import (
	"io"

	"example.com/extlib"
	"svc/domain/model"
	"svc/server/grpc"
)

// LocalRepository — интерфейс, объявленный в этом же пакете рядом с
// потребителем. Использование — норма.
type LocalRepository interface {
	Job(id string) (model.Job, error)
}

// --- Позитивный класс: интерфейс из чужого «своего» пакета сервиса ---

// Поле структуры: интерфейс из чужого server-пакета.
type Service struct {
	notifier grpc.Notifier // want `GID-134: интерфейс Notifier объявлен в svc/server/grpc — определите интерфейс рядом с потребителем \(исключения: библиотеки и /domain/model для service/usecase\)`
	local    LocalRepository
}

// Параметр функции: интерфейс из чужого server-пакета.
func (s *Service) Register(n grpc.Notifier) {} // want `GID-134: интерфейс Notifier объявлен в svc/server/grpc`

// Результат функции: интерфейс из чужого server-пакета.
func (s *Service) Notifier() grpc.Notifier { return nil } // want `GID-134: интерфейс Notifier объявлен в svc/server/grpc`

// --- Негативный класс: чистый код ---

// Интерфейс из model-слоя у потребителя service — ОК.
func (s *Service) WithRepo(r model.JobRepository) {}

// Интерфейс из того же пакета — ОК.
func (s *Service) WithLocal(l LocalRepository) {}

// Библиотечный интерфейс stdlib (io.Reader) — ОК.
func (s *Service) Read(r io.Reader) {}

// Интерфейс внешней библиотеки — ОК.
func (s *Service) Encode(e extlib.Encoder) {}

// --- Класс неприменимости ---

// error — не задевается (нет пакета объявления).
func (s *Service) Do() error { return nil }

// Анонимный интерфейс — не именованный, не задевается.
func (s *Service) Anon(x interface{ Foo() }) {}

// any / interface{} — не задевается.
func (s *Service) Any(v any) {}

// Не-интерфейсные типы (struct, string) — не задеваются.
func (s *Service) Plain(j model.Job, name string) {}
