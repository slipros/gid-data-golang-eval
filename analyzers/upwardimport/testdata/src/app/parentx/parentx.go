// Граничный: "app/parentx" НЕ является дочерним для "app/parent" — префикс
// считается по сегментам пути, а не по строке. Строка "app/parent" является
// строковым префиксом "app/parentx", но НЕ сегментным (нет "app/parent/").
// Диагностики быть не должно.
package parentx

import "app/parent" // ок: parentx не дочерний для parent (префикс по сегментам)

// Holder использует тип из не-родительского пакета parent.
type Holder struct {
	Root parent.Root
}
