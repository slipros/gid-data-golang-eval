// Eval GID-230: non-applicability — "handler" segment without a "server"
// segment in the import path: the rule does not apply.
package handler

// Job has no Handle method but lives outside the /server layer.
type Job struct{ ID string }
