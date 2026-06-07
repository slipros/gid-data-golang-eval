// Eval для GID-154 (logger WithField в конструкторе).
package logconstruct

import "github.com/sirupsen/logrus"

// --- Позитив: конструктор не вызывает WithField ---

type Snapshot struct {
	logger *logrus.Entry
}

func NewSnapshot(logger *logrus.Entry) *Snapshot { // want `GID-154: сущность "Snapshot" содержит logger — конструктор "NewSnapshot" обязан вызвать logger\.WithField\(<entity>, <name>\)`
	return &Snapshot{logger: logger}
}

// --- Негатив: WithField вызван ---

type Job struct {
	logger *logrus.Entry
}

func NewJob(logger *logrus.Entry) *Job {
	return &Job{logger: logger.WithField("service", "job")}
}

// Граничный кейс: logger-интерфейс logrus.FieldLogger.
type Upload struct {
	logger logrus.FieldLogger
}

func NewUpload(logger logrus.FieldLogger) *Upload {
	return &Upload{logger: logger.WithField("service", "upload")}
}

// --- Неприменимость: сущность без logger ---

type Plain struct {
	retries int
}

func NewPlain(retries int) *Plain {
	return &Plain{retries: retries}
}
