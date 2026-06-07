// Package pathseg — определение слоя Clean Architecture по сегментам
// import-пути пакета. Конвенция: слой задаётся последовательностью
// сегментов, например /dal/repository или /domain/model — независимо от
// префикса модуля.
package pathseg

import "strings"

// Index возвращает индекс первого вхождения seq как последовательных
// сегментов пути, либо -1.
func Index(path string, seq ...string) int {
	segs := Segments(path)
	if len(seq) == 0 || len(segs) < len(seq) {
		return -1
	}
	for i := 0; i+len(seq) <= len(segs); i++ {
		if matchAt(segs, i, seq) {
			return i
		}
	}
	return -1
}

// Contains сообщает, содержит ли путь seq как последовательные сегменты.
func Contains(path string, seq ...string) bool {
	return Index(path, seq...) >= 0
}

// EndsWith сообщает, заканчивается ли путь сегментами seq —
// т.е. пакет является корнем слоя, а не его подпакетом.
func EndsWith(path string, seq ...string) bool {
	segs := Segments(path)
	if len(segs) < len(seq) {
		return false
	}
	return matchAt(segs, len(segs)-len(seq), seq)
}

// Segments разбивает import-путь на сегменты.
func Segments(path string) []string {
	return strings.Split(path, "/")
}

func matchAt(segs []string, i int, seq []string) bool {
	for j, s := range seq {
		if segs[i+j] != s {
			return false
		}
	}
	return true
}
