// Eval for GID-270 (model-place), part B: a package outside /domain/** is
// out of scope.
package http

// Request — a transport package: not /domain/service, not /domain/usecase,
// not convert.
type Request struct {
	Query string
}

// Decode — non-applicability: /server/http is neither convert nor a domain
// layer, so its own structs stay put.
func Decode(query string) Request {
	return Request{Query: query}
}
