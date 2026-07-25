// Eval for GID-154 (logger WithField in the constructor).
package logconstruct

import "github.com/sirupsen/logrus"

// --- Positive: the constructor does not call WithField ---

type Snapshot struct {
	logger *logrus.Entry
}

func NewSnapshot(logger *logrus.Entry) *Snapshot { // want `GID-154: entity "Snapshot" has a logger\. Fix: constructor "NewSnapshot" must name the entity on it`
	return &Snapshot{logger: logger}
}

// --- Negative: WithField is called ---

type Job struct {
	logger *logrus.Entry
}

func NewJob(logger *logrus.Entry) *Job {
	return &Job{logger: logger.WithField("service", "job")}
}

// Boundary case: the logrus.FieldLogger logger interface.
type Upload struct {
	logger logrus.FieldLogger
}

func NewUpload(logger logrus.FieldLogger) *Upload {
	return &Upload{logger: logger.WithField("service", "upload")}
}

// --- Non-applicability: an entity without a logger ---

type Plain struct {
	retries int
}

func NewPlain(retries int) *Plain {
	return &Plain{retries: retries}
}

// --- Positive: a CamelCase entity name in the logrus shapes ---

type snapshotStore struct {
	logger *logrus.Entry
}

func NewSnapshotStore(logger *logrus.Entry) *snapshotStore {
	return &snapshotStore{logger: logger.WithField("service", "SnapshotStore")} // want `GID-154: the entity name "SnapshotStore" is not lower snake_case\. Fix: spell it as the log fields do`
}

func NewSnapshotStoreFields(logger *logrus.Entry) *snapshotStore {
	return &snapshotStore{logger: logger.WithFields(logrus.Fields{"service": "snapshotStore"})} // want `GID-154: the entity name "snapshotStore" is not lower snake_case\. Fix: spell it as the log fields do`
}

// --- Negative: lower snake_case in both shapes ---

func NewSnapshotStoreGood(logger *logrus.Entry) *logrus.Entry {
	return logger.WithField("service", "snapshot_store")
}
