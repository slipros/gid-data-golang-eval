.PHONY: build eval lint lint-fast install-hook clean

## build: собрать кастомный бинарь custom-gcl (требует golangci-lint v2.9.0)
build:
	golangci-lint custom

## eval: прогнать eval всех правил (analysistest + testdata)
eval:
	go test ./...

## lint: пересобрать custom-gcl и прогнать на репозитории
lint: build lint-fast

## lint-fast: прогнать уже собранный custom-gcl (без пересборки)
lint-fast:
	./bin/custom-gcl run ./...

## install-hook: поставить git pre-commit hook с локальной проверкой
install-hook:
	@printf '#!/bin/sh\n# Локальный гейт правил (см. RULES.md).\nexec make -s lint-fast\n' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "pre-commit hook установлен (.git/hooks/pre-commit)"

clean:
	rm -rf bin
