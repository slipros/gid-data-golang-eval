// Eval GID-213: not applicable — an ordinary package outside the validate layer.
package service

// A struct without a Validate method in /domain/service — the rule does not apply.
type Worker struct{ ID string }
