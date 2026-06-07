// Eval для GID-197: settings.exclude — "Интерфейс" и "Интерфейс.Метод".
package service

type LegacyGateway interface {
	Send(msg string) // интерфейс исключён целиком
}

type AlertSink interface {
	Alert(msg string) // want `GID-197: method "Alert" of interface "AlertSink" is not used in the consumer package\. Fix: keep the interface minimal, remove the method`
	Flush() error     // исключён как AlertSink.Flush
}
