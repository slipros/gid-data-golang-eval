// Eval boundary GID-212: a dal/repository/build segment nested under another
// layer (client/x) is NOT the build layer — the layer is anchored to the
// module root (pathseg.HasLayer), so an exported function with a signature
// that does not match the build contract here must NOT be flagged, unlike
// dalsvc/dal/repository/build itself.
package build

// Would be flagged by the signature-contract check if the layer segment were
// matched anywhere in the path (pathseg.Contains) instead of being anchored
// to the module root.
func DoSomething(id string) (int, error) { return 0, nil }
