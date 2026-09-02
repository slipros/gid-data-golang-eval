// Package aliased — the diagnostic spells a type the way the file spells it.
package aliased

import (
	ucase "svc/usecase"
)

// The import carries an alias and the assertion names the interface through it,
// so the diagnostic quotes the line the way the source writes it.
var _ ucase.LatestPageStore = objectStore{} // want `GID\-274:\ redundant\ compile\-time\ assertion:\ the\ package\ already\ passes\ this\ value\ as\ ucase\.LatestPageStore\ at\ aliased\.go:18,\ so\ the\ compiler\ checks\ the\ contract\ there\.\ Fix:\ delete\ the\ "var\ _\ ucase\.LatestPageStore\ =\ objectStore\{\}"\ line`

type objectStore struct{}

func (objectStore) Object(key string) []byte { return []byte(key) }

// Wire hands the store over to the page.
func Wire() *ucase.LatestPage {
	return ucase.NewLatestPage(nil, nil, objectStore{}, nil)
}
