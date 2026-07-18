// Boundary case: a package literally named dal/entity but nested under
// another layer (server/grpc) is NOT the entity layer — pathseg.HasLayer
// anchors the layer to the module root, so this package must not be treated
// as /dal/entity (would false-match under a plain path Contains).
package entity

type FakeEntity struct {
	ID string
}
