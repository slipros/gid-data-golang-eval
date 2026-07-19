// Eval GID-143: the missing-key handler is matched by symbol name only, so a
// helper library on a NON-gitlab module path (git.k8s.nomilk.space/...) counts
// as valid handling. This whole package must produce NO diagnostics.
package convert

import gderror "git.k8s.nomilk.space/go-library/ehelper"

// An enum per GID-123: a named type based on string.
type (
	EntityFormat string
	ModelFormat  string
)

var formatMap = map[EntityFormat]ModelFormat{
	"wav": "wav",
}

// comma-ok + handling via NewUnhandledValueError from the nomilk-path helper.
// The path differs from the historical gitlab one — must still be recognized.
func ModelFormatFromEntity(f EntityFormat) (ModelFormat, error) {
	v, ok := formatMap[f]
	if !ok {
		return "", gderror.NewUnhandledValueError(f)
	}
	return v, nil
}
