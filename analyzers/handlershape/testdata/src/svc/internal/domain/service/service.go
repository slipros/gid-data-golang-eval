// Eval GID-230: non-applicability — package outside the /server layer.
package service

// UnimplementedFooServer — looks like a proto embed but lives outside /server.
type UnimplementedFooServer struct{}

// Worker embeds Unimplemented*Server and has non-Handler fields, but the
// package is not under the "server" segment — the rule does not apply.
type Worker struct {
	UnimplementedFooServer

	queue string
}
