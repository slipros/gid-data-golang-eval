// Eval GID-212 (boundary): the squirrel ban still applies in _test.go files
// outside build — only the signature check is out of scope for tests.
package repository

import (
	"testing"

	"github.com/Masterminds/squirrel" // want `GID-212: squirrel is allowed only in repository build packages \(/dal/repository/build\)\. Fix: move squirrel usage into /dal/repository/build`
)

func TestDoStuff(t *testing.T) {
	if _, _, err := squirrel.Select("id").ToSql(); err != nil {
		t.Fatal(err)
	}
}
