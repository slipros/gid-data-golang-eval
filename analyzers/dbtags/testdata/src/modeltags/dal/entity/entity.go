// Неприменимость: та же структура в /dal/entity — db-теги тут обязательны
// (GID-125), GID-168 здесь не действует.
package entity

import "time"

type Snapshot struct {
	ID        string    `db:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at"`
}
