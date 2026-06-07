// Eval для GID-197: settings.exclude — "Интерфейс" и "Интерфейс.Метод".
package service

type LegacyGateway interface {
	Send(msg string) // интерфейс исключён целиком
}

type AlertSink interface {
	Alert(msg string) // want `GID-197: метод "Alert" интерфейса "AlertSink" не используется в пакете-потребителе — интерфейс минимален: уберите метод из интерфейса`
	Flush() error     // исключён как AlertSink.Flush
}
