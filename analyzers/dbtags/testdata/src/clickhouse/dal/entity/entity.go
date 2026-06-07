// Eval settings.tags: ClickHouse-библиотека маппит тегом ch.
package entity

// Негатив: ch-тег принят наравне с db.
type Event struct {
	ID   string `ch:"id"`
	Name string `db:"name"`
}

// Позитив: нет ни одного допустимого тега.
type Metric struct {
	Value float64 `json:"value"` // want `GID-125: поле Metric\.Value без тега маппинга \(db/ch\) — соответствие entity колонкам БД явное`
}
