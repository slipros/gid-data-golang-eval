// Boundary: a package literally named "server" nested under a client SDK
// (client/webhook/server) is NOT the server Clean-Architecture layer —
// pathseg.HasLayer anchors a layer to the module root, unlike pathseg.Contains
// which would match "server" anywhere in the path. Without the fix this
// converter-shaped function outside convert/ would falsely trigger GID-135;
// with the fix it does not, since the leading path segment is "client", not
// "server".
package server

type Event struct{ ID string }

type Request struct{ ID string }

func EventPayloadFromRequest(in *Request) Event {
	return Event{ID: in.ID}
}
