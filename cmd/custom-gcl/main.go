// Команда custom-gcl — полноценный golangci-lint v2.9.0 со встроенными
// gid*-линтерами этого репозитория.
//
// Это альтернатива сборке через `golangci-lint custom` (.custom-gcl.yml):
// бинарь ставится напрямую и не требует клонирования golangci-lint —
//
//	go install github.com/slipros/gid-data-golang-eval/cmd/custom-gcl@latest
//
// Запуск идентичен обычному golangci-lint (нужен .golangci.yml с включёнными
// gid*-линтерами и ruleguard/rules.go рядом):
//
//	custom-gcl run ./...
//
// Версия golangci-lint фиксирована в go.mod (v2.9.0) — должна совпадать
// с версией из .custom-gcl.yml.
package main

import (
	"fmt"
	"os"

	"github.com/golangci/golangci-lint/v2/pkg/commands"
	"github.com/golangci/golangci-lint/v2/pkg/exitcodes"

	// Регистрирует все gid*-линтеры через init() пакета gidrules.
	// Точка сборки бинаря обязана импортировать корневой пакет —
	// тот же контракт, что у сгенерированного `golangci-lint custom`.
	//nolint:gidupwardimport // composition root плагина импортирует корень по контракту plugin system
	_ "github.com/slipros/gid-data-golang-eval"
)

func main() {
	info := commands.BuildInfo{
		Version:   "custom-gcl (gid-data-golang-eval)",
		Commit:    "(see go module version)",
		Date:      "(unknown)",
		GoVersion: "unknown",
	}
	if err := commands.Execute(info); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "The command is terminated due to an error: %v\n", err)
		os.Exit(exitcodes.Failure)
	}
}
