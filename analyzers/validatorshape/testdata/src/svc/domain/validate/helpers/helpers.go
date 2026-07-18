// Eval GID-213: boundary — a subpackage nested under a validate package
// (domain/validate/helpers) is not itself the validate package: "validate"
// is not the trailing (leaf) segment here. Without the EndsWith fix,
// pathseg.Contains would have falsely put this package in scope and flagged
// CreateJob below.
package helpers

// Boundary class: an exported struct without a Validate method, declared in
// a package that is merely nested under validate/, not the validate leaf
// package itself — the rule does not apply.
type CreateJob struct{}
