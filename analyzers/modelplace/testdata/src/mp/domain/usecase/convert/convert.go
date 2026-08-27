// Eval for GID-270 (model-place), part A: a convert package nested under
// /domain is judged by part A only — one diagnostic, not two.
package convert

// Send — an exported struct in usecase/convert: part A reports it once.
type Send struct { // want `GID-270: type "Send" is declared in a convert package`
	MessageID string
	Body      string
}
