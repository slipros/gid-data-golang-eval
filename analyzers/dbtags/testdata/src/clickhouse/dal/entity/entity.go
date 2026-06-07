// Eval settings.tags: ClickHouse-библиотека маппит тегом ch.
package entity

// Негатив: ch-тег принят наравне с db.
type Event struct {
	ID   string `ch:"id"`
	Name string `db:"name"`
}

// Позитив: нет ни одного допустимого тега.
type Metric struct {
	Value float64 `json:"value"` // want `GID-125: field Metric\.Value has no mapping tag \(db/ch\)\. Fix: add a tag so entity-to-column mapping is explicit`
}
