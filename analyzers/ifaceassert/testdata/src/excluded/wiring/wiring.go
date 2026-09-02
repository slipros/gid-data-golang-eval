// Package wiring — the composition root judged with settings.exclude.
package wiring

// Store — the port kept on the exclusion list.
type Store interface {
	Object(key string) []byte
}

// Parquet — the port judged as usual.
type Parquet interface {
	Rows(data []byte) int
}

// On settings.exclude by the interface name: the assertion stays.
var _ Store = objectStore{}

// On settings.exclude qualified by the type: the assertion stays.
var _ Parquet = rowReader{}

// Judged as usual — the same interface, another type.
var _ Parquet = otherReader{} // want `GID\-274:\ redundant\ compile\-time\ assertion:\ the\ package\ already\ passes\ this\ value\ as\ Parquet\ at\ wiring\.go:42,\ so\ the\ compiler\ checks\ the\ contract\ there\.\ Fix:\ delete\ the\ "var\ _\ Parquet\ =\ otherReader\{\}"\ line`

type objectStore struct{}

func (objectStore) Object(key string) []byte { return []byte(key) }

type rowReader struct{}

func (rowReader) Rows(data []byte) int { return len(data) }

type otherReader struct{}

func (otherReader) Rows(data []byte) int { return len(data) }

// UsePage takes both ports.
func UsePage(store Store, parquet Parquet) int { return parquet.Rows(store.Object("k")) }

// Wire hands every implementation over.
func Wire() int {
	_ = UsePage(objectStore{}, rowReader{})

	return UsePage(objectStore{}, otherReader{})
}
