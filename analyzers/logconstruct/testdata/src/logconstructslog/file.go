// Eval for GID-154 on the slog stack: the entity is named on the logger with
// With("<entity>", <name>) instead of logrus' WithField.
package logconstructslog

import "log/slog"

// --- Positive: the constructor does not name the entity on the logger ---

type Snapshot struct {
	logger *slog.Logger
}

func NewSnapshot(logger *slog.Logger) *Snapshot { // want `GID-154: entity "Snapshot" has a logger\. Fix: constructor "NewSnapshot" must name the entity on it`
	return &Snapshot{logger: logger}
}

// --- Negative: the entity is named with With ---

type Job struct {
	logger *slog.Logger
}

func NewJob(logger *slog.Logger) *Job {
	return &Job{logger: logger.With("job", "sweep")}
}

// Boundary: an attribute-typed With is the same shape.
type Task struct {
	logger *slog.Logger
}

func NewTask(logger *slog.Logger) *Task {
	return &Task{logger: logger.With(slog.String("task", "encode"))}
}

// --- Non-applicability: an entity without a logger ---

type Plain struct {
	name string
}

func NewPlain(name string) *Plain {
	return &Plain{name: name}
}

// --- Boundary: the constructor takes a logger but stores it in an entity
// whose name does not match the constructor — the requirement still holds ---

type reporter struct {
	logger *slog.Logger
}

func NewReporter(logger *slog.Logger) *reporter { // want `GID-154: entity "Reporter" has a logger\. Fix: constructor "NewReporter" must name the entity on it`
	return &reporter{logger: logger}
}

// Negative: the same shape, with the entity named on the logger.
func NewReporterNamed(logger *slog.Logger) *reporter {
	return &reporter{logger: logger.With(slog.String("reporter", "daily"))}
}

// Boundary: a constructor that passes the logger on to another constructor
// must still name its own entity first.
type wrapper struct {
	inner *reporter
}

func NewWrapper(logger *slog.Logger) *wrapper {
	return &wrapper{inner: NewReporterNamed(logger.With(slog.String("wrapper", "outer")))}
}
