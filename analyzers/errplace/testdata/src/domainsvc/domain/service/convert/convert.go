// Negative (boundary): typed gderror errors are allowed —
// converters use NewUnhandledValueError for enum map mapping.
package convert

import gderror "gitlab.gid.team/gid-data/tech/golang/libs/helper.git/errors"

func StatusFromEntity(in string) (string, error) {
	m := map[string]string{"pending": "pending"}
	out, ok := m[in]
	if !ok {
		return "", gderror.NewUnhandledValueError(in)
	}
	return out, nil
}
