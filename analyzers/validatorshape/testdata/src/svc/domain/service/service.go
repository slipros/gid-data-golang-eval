// Eval GID-213: неприменимость — обычный пакет вне слоя validate.
package service

// Struct без метода Validate в /domain/service — правило не применяется.
type Worker struct{ ID string }
