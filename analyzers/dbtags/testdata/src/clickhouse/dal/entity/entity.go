// Eval settings.tags: the ClickHouse library maps with the ch tag.
package entity

// Negative: the ch tag is accepted on par with db.
type Event struct {
	ID   string `ch:"id"`
	Name string `db:"name"`
}

// Positive: not a single allowed tag.
type Metric struct {
	Value float64 `json:"value"` // want `GID-125: field Metric\.Value has no mapping tag \(db/ch\)\. Fix: add a tag so entity-to-column mapping is explicit`
}
