// Сосед для проверки негатива: импорт соседнего пакета (app/parent/other)
// из app/parent/badchild не является импортом родителя.
package other

// Sibling — тип соседнего дочернего пакета.
type Sibling struct {
	ID int
}
