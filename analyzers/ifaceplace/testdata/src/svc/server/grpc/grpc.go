// Package grpc — «свой» пакет сервиса (слой /server/grpc), НЕ model и НЕ
// тот же пакет, что у потребителя-репозитория/сервиса. Интерфейсы отсюда
// использовать в чужих пакетах нельзя — определяй рядом с потребителем.
package grpc

// Notifier — интерфейс чужого пакета сервиса (server-слой).
type Notifier interface {
	Notify(msg string) error
}
