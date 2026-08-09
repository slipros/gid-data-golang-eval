package foreign

// Router — a router of some other library entirely.
type Router any

// RouterFunc — its registrar type.
type RouterFunc = func(r Router) error

// Server — its server.
type Server struct{}

// NewServer — a same-named constructor from a foreign package: the rule matches
// the library by import path, so nothing here is judged.
func NewServer(addr string, routers ...RouterFunc) *Server {
	_, _ = addr, routers

	return &Server{}
}

// routes — the application's own routes.
func routes(r Router) error {
	_ = r

	return nil
}

// Build hands a bare router to a foreign server: not this rule's business.
func Build() *Server {
	return NewServer(":8080", routes)
}
