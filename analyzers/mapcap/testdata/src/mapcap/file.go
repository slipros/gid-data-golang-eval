// Eval для GID-183 (map capacity hint при заполнении из range).
package mapcap

// --- Класс 1: позитивные (make(map) без cap + безусловное заполнение в range) ---

// Заполнение из range по слайсу.
func fillFromSlice(src []int) map[int]int {
	m := make(map[int]int) // want `GID-183: make without capacity while filling from range\. Fix: make\(map\[K\]V, len\(src\)\)`
	for _, v := range src {
		m[v] = v
	}
	return m
}

// var-форма объявления.
func fillFromSliceVar(src []string) map[string]bool {
	var m = make(map[string]bool) // want `GID-183: make without capacity while filling from range\. Fix: make\(map\[K\]V, len\(src\)\)`
	for _, v := range src {
		m[v] = true
	}
	return m
}

// Заполнение из range по мапе.
func fillFromMap(src map[string]int) map[string]int {
	m := make(map[string]int) // want `GID-183: make without capacity while filling from range\. Fix: make\(map\[K\]V, len\(src\)\)`
	for k, v := range src {
		m[k] = v
	}
	return m
}

// Заполнение из range по строке.
func fillFromString(src string) map[rune]int {
	m := make(map[rune]int) // want `GID-183: make without capacity while filling from range\. Fix: make\(map\[K\]V, len\(src\)\)`
	for _, r := range src {
		m[r] = 1
	}
	return m
}

// --- Класс 2: негативные ---

// make с уже указанной ёмкостью — корректно.
func withCapacity(src []int) map[int]int {
	m := make(map[int]int, len(src))
	for _, v := range src {
		m[v] = v
	}
	return m
}

// Заполнение без range — размер не выводится из коллекции.
func fillWithoutRange() map[int]int {
	m := make(map[int]int)
	m[1] = 1
	m[2] = 2
	return m
}

// range по каналу — длина неизвестна, размер не подсказать.
func fillFromChan(src chan int) map[int]int {
	m := make(map[int]int)
	for v := range src {
		m[v] = v
	}
	return m
}

// --- Класс 3: граничные (не матчатся) ---

// Условное заполнение внутри if в теле цикла — реальный размер < len(src).
func conditionalFill(src []int) map[int]int {
	m := make(map[int]int)
	for _, v := range src {
		if v > 0 {
			m[v] = v
		}
	}
	return m
}

// m используется между make и циклом (заполнение вне цикла) — отменяет диагностику.
func usedBeforeLoop(src []int) map[int]int {
	m := make(map[int]int)
	m[0] = 0
	for _, v := range src {
		m[v] = v
	}
	return m
}

// m передаётся в вызов между make и циклом — отменяет диагностику.
func passedBeforeLoop(src []int) map[int]int {
	m := make(map[int]int)
	consume(m)
	for _, v := range src {
		m[v] = v
	}
	return m
}

func consume(m map[int]int) { _ = m }

// --- Класс 4: неприменимость (нет make(map)) ---

// make слайса — не мапа.
func makeSlice(src []int) []int {
	s := make([]int, 0)
	for _, v := range src {
		s = append(s, v)
	}
	return s
}
