// Eval для GID-004 (allptr).
package allptr

import gdhelper "gitlab.gid.team/gid-data/tech/golang/libs/helper.git"

type File struct {
	ID   string
	Name string
}

type Files []File

// --- Позитивные кейсы: нарушение ловится ---

func bad(files []File) []string {
	var out []string
	for _, f := range files { // want `GID-004: итерация по слайсу структур — используйте gdhelper\.AllPtr`
		out = append(out, f.Name)
	}
	return out
}

// Граничный кейс: именованный слайс-тип.
func badNamed(files Files) []string {
	var out []string
	for _, f := range files { // want `GID-004: итерация по слайсу структур — используйте gdhelper\.AllPtr`
		out = append(out, f.Name)
	}
	return out
}

// Граничный кейс: итерация только по индексу — тоже нарушение,
// стайлгайд требует AllPtr вместо любых range-форм по слайсу структур.
func badIndexOnly(files []File) int {
	n := 0
	for i := range files { // want `GID-004: итерация по слайсу структур — используйте gdhelper\.AllPtr`
		n += i
	}
	return n
}

// --- Негативные кейсы: чистый код проходит ---

func good(files []File) []string {
	var out []string
	for _, f := range gdhelper.AllPtr(files) {
		out = append(out, f.Name)
	}
	return out
}

// Слайс указателей — копирования нет, AllPtr не нужен.
func goodPtrSlice(files []*File) []string {
	var out []string
	for _, f := range files {
		out = append(out, f.Name)
	}
	return out
}

// --- Неприменимость: не слайсы структур ---

func notApplicableStrings(names []string) int {
	n := 0
	for range names {
		n++
	}
	return n
}

func notApplicableMap(byID map[string]File) []string {
	var out []string
	for id := range byID {
		out = append(out, id)
	}
	return out
}
