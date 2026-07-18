// Boundary case: a package literally containing the domain/model segments
// but nested under another layer (server/grpc) is NOT the model layer —
// pathseg.HasLayer anchors the layer to the module root, so this must NOT
// be flagged (would be a false positive under a plain path Contains).
package nested

type FakeAlias = string
