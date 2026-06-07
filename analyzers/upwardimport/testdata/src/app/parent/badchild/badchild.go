// Позитив + негатив:
//   - импорт родителя "app/parent" из дочернего пакета — нарушение GID-131;
//   - импорт соседа "app/parent/other" — НЕ родитель, диагностики нет.
package badchild

import (
	"app/parent"       // want `GID-131: дочерний пакет импортирует родительский app/parent — инвертируйте зависимость: общее выносится вниз, родитель импортирует детей`
	"app/parent/other" // ок: сосед, а не родитель
)

// Up тянет тип из родительского и соседнего пакетов.
type Up struct {
	Root    parent.Root
	Sibling other.Sibling
}
