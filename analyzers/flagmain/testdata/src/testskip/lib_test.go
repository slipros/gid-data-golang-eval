// Файл *_test.go: flag в тестах легитимен — анализатор такие файлы
// пропускает, диагностики нет даже в не-main пакете и при не-snake_case имени.
package testskip

import (
	"flag"
	"testing"
)

var update = flag.Bool("updateGolden", false, "update golden files")

func TestAdd(t *testing.T) {
	flag.Parse()
	if Add(1, 2) != 3 {
		t.Fatal("bad")
	}
	_ = update
}
