// Класс 4: неприменимость — пакет без импорта забаненной библиотеки.
package clean

import "example.com/otherdb"

// Никаких запрещённых символов: чисто.
func loadOther() (int, error) {
	return otherdb.TQuery[int]("select 1")
}

func plain(x int) int { return x * 2 }
