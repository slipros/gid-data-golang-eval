// Boundary case: domain/model nested under another layer (server/grpc) is
// NOT the model layer — pathseg.HasLayer anchors the layer to the module
// root, so an exported package-level constant here must be flagged as
// out-of-scope (would be wrongly allowed and skipped under a plain path
// Contains, since "domain", "model" still occurs mid-path).
package nested

const ExportedConst = "value" // want `GID-194: exported constant "ExportedConst" is declared outside model/entity\. Fix: keep shared constants in /domain/model or /dal/entity, and declare local ones where they are used`
