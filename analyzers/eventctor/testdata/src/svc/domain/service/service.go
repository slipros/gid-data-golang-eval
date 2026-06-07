// Eval для GID-216: вне event-слоя правило не применяется (неприменимость).
package service

type Service struct{}

// Конструктор без logger вне event-слоя — НЕ флагается.
func NewService() *Service {
	return &Service{}
}
