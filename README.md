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

1. В репозитории сервиса завести `.custom-gcl.yml`:

   ```yaml
   version: v2.9.0
   name: custom-gcl
   destination: ./bin
   plugins:
     - module: 'github.com/slipros/gid-data-golang-eval'
       version: <тег>          # или path: /локальный/путь для разработки
   ```

2. Собрать бинарь: `golangci-lint custom`.
3. Взять за основу эталонный [.golangci.yml](.golangci.yml) — включить нужные
   `gid*`-линтеры, настроить исключения (`settings.exclude`, `settings.tree`,
   `settings.tags`, …) и скопировать `ruleguard/rules.go` (путь задаётся в
   `settings.gocritic.settings.ruleguard.rules`).
4. Запуск: `./bin/custom-gcl run ./...`.

## IDE

Чтобы диагностики были видны прямо в редакторе, IDE должна вызывать
`bin/custom-gcl` вместо обычного golangci-lint:

- **VS Code** (`settings.json`):

  ```json
  {
    "go.lintTool": "golangci-lint-v2",
    "go.alternateTools": { "golangci-lint-v2": "${workspaceFolder}/bin/custom-gcl" }
  }
  ```

- **GoLand**: Settings → Tools → Go Linter (плагин golangci-lint) → указать
  путь к `bin/custom-gcl`.

## Исключения из правил

Два уровня (подробности в [RULES.md](RULES.md)):

- точечно — `//nolint:<линтер>` с комментарием-обоснованием;
- централизованно — `settings` линтера в `.golangci.yml`
  (например, `gidcreateupdate.settings.exclude`, `giddbtags.settings.tags`,
  `giddirtree.settings.tree`).

## Добавление нового правила

Процесс — в конце [RULES.md](RULES.md): строка реестра → реализация →
**обязательный eval** (analysistest, 4 класса кейсов) → включение в конфиг.
