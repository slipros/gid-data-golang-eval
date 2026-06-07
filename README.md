# gid-data-golang-eval

Кастомный плагин golangci-lint, переносящий правила внутреннего стайлгайда
(skill `go-styleguide`) в детерминированный линтер для **локальной разработки**.

## Статус: линтер полностью заменяет стилевую часть скилла `go-styleguide`

Верифицировано 2026-06-07 полной сверкой всех доков скилла (31 файл) с реестром:

- каждое детерминируемое правило стайлгайда реализовано (GID-001…GID-217, все со
  статусом ✅ и обязательным eval) либо покрыто стандартными линтерами golangci-lint
  (слой 3, GID-201…GID-209);
- непереносимые эвристики явно перечислены в [RULES.md](RULES.md) «Не переносим»
  и [FINDINGS.md](FINDINGS.md) §2.4/2.5 — остаются на code review осознанно;
- сверх скилла добавлены правила Uber/Google best practices (GID-178…GID-197),
  которых скилл не проверял, — итоговый контроль стиля строже ручного review по скиллу;
- `make eval` — все analysistest зелёные; `make lint-fast` на самом репозитории —
  0 issues.

Скилл остаётся источником шаблонов кода и формата проектной документации
(спеки задач, README-индексы); проверку стиля кода он больше не выполняет —
это делает линтер детерминированно.

- **[RULES.md](RULES.md)** — реестр правил со статусами; у каждого правила обязателен eval
- **[PRD.md](PRD.md)** — концепция
- `analyzers/` — собственные go/analysis-анализаторы (одно правило = один линтер)
- `ruleguard/rules.go` — простые паттерн-правила (gocritic → ruleguard)
- `.golangci.yml` — эталонный конфиг со всеми линтерами и примерами настроек;
  база — боевой конфиг сервиса consent-api (UDMP/backend-go), поверх — слои GID

## Быстрый старт

Требуется golangci-lint **v2.9.0** (версия зафиксирована в `.custom-gcl.yml`).

```sh
make build         # собрать бинарь bin/custom-gcl
make eval          # прогнать eval всех правил (go test ./...)
make lint-fast     # проверить код собранным бинарём
make install-hook  # git pre-commit hook с локальной проверкой
```

## Подключение в своём сервисе

`gid*`-линтеры — это module-плагины golangci-lint: обычный `golangci-lint run`
их **не видит**, они вкомпилируются в отдельный бинарь `custom-gcl` (полный
golangci-lint v2.9.0 + наши линтеры). Собранным бинарём пользуешься как обычным
golangci-lint — стандартные и `gid*`-линтеры работают одним прогоном по одному
`.golangci.yml`. Бинарь собирается одним из двух способов.

### Способ A — `go install` (рекомендуется)

Бинарь ставится напрямую, без клонирования golangci-lint:

```sh
go install github.com/slipros/gid-data-golang-eval/cmd/custom-gcl@v0.3.0
```

`custom-gcl` появится в `$(go env GOPATH)/bin` (добавь в `PATH`). Обновление —
тем же `go install` с новым тегом. В каждом сервисе нужны только `.golangci.yml`
и `ruleguard/rules.go` — клонировать ничего не нужно.

### Способ B — `golangci-lint custom` (.custom-gcl.yml)

Локальный бинарь в проекте (нужен установленный golangci-lint v2.9.0):

```yaml
# .custom-gcl.yml
version: v2.9.0
name: custom-gcl
destination: ./bin
plugins:
  - module: 'github.com/slipros/gid-data-golang-eval'
    version: v0.3.0          # или path: /локальный/путь для разработки
```

Собрать: `golangci-lint custom` → `./bin/custom-gcl`.

### Дальше (для обоих способов)

1. Взять за основу эталонный [.golangci.yml](.golangci.yml) — включить нужные
   `gid*`-линтеры, настроить исключения (`settings.exclude`, `settings.tree`,
   `settings.tags`, …); убрать репо-специфичные куски (exclusions для `testdata`
   и `ruleguard/rules.go`, `giddirtree.settings.tree`).
2. Скопировать `ruleguard/rules.go` в сервис (путь задаётся в
   `settings.gocritic.settings.ruleguard.rules`, по умолчанию `${base-path}/ruleguard/rules.go`).
3. Запуск: `custom-gcl run ./...` (способ A) или `./bin/custom-gcl run ./...` (способ B).

## IDE

Чтобы диагностики были видны прямо в редакторе, IDE должна вызывать
`custom-gcl` вместо обычного golangci-lint. Путь — `$(go env GOPATH)/bin/custom-gcl`
при `go install` (способ A) либо `${workspaceFolder}/bin/custom-gcl` при сборке
в проект (способ B):

- **VS Code** (`settings.json`):

  ```json
  {
    "go.lintTool": "golangci-lint-v2",
    "go.alternateTools": { "golangci-lint-v2": "custom-gcl" }
  }
  ```

  (`custom-gcl` из `PATH` при `go install`; иначе абсолютный путь к бинарю.)

- **GoLand**: Settings → Tools → Go Linter (плагин golangci-lint) → указать
  путь к `custom-gcl`.

## Исключения из правил

Два уровня (подробности в [RULES.md](RULES.md)):

- точечно — `//nolint:<линтер>` с комментарием-обоснованием;
- централизованно — `settings` линтера в `.golangci.yml`
  (например, `gidcreateupdate.settings.exclude`, `giddbtags.settings.tags`,
  `giddirtree.settings.tree`).

## Добавление нового правила

Процесс — в конце [RULES.md](RULES.md): строка реестра → реализация →
**обязательный eval** (analysistest, 4 класса кейсов) → включение в конфиг.
