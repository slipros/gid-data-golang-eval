// Eval GID-212: the external test package of a build package (build_test) counts
// as a build package for the squirrel ban — no diagnostic on the import.
package build_test

import (
	"testing"

	"github.com/Masterminds/squirrel"
)

func TestExpectedSQL(t *testing.T) {
	sql, _, err := squirrel.Select("id").ToSql()
	if err != nil || sql == "" {
		t.Fatal("bad expectation")
	}
}
