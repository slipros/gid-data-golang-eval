// Позитив + негатив:
//   - импорт родителя "app/parent" из дочернего пакета — нарушение GID-131;
//   - импорт соседа "app/parent/other" — НЕ родитель, диагностики нет.
package badchild

import (
	"app/parent"       // want `GID-131: a child package imports its parent app/parent\. Fix: invert the dependency, move shared code down and let the parent import children`
	"app/parent/other" // ок: сосед, а не родитель
)

// Up тянет тип из родительского и соседнего пакетов.
type Up struct {
	Root    parent.Root
	Sibling other.Sibling
}
