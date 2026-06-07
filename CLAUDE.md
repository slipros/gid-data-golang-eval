# CLAUDE.md

Кастомный плагин golangci-lint (module plugin system), переносящий правила внутреннего
стайлгайда gid.team (skill `go-styleguide`) в детерминированный линтер. Каждое правило
имеет ID `GID-NNN` и зарегистрировано в [RULES.md](RULES.md).

## Команды

```bash
make build         # собрать бинарь bin/custom-gcl (golangci-lint custom)
make eval          # прогнать eval всех правил (go test ./...)
make lint-fast     # проверить код репозитория собранным бинарём
go test ./analyzers/<slug>/...   # eval одного правила
```

Сборка требует golangci-lint **v2.9.0** — версия зафиксирована в `.custom-gcl.yml`.
Версии зависимостей пинятся под golangci v2.9.0 — не обновлять без проверки сборки.

## Структура

- `analyzers/<slug>/` — go/analysis-анализаторы: одно правило (или группа смежных GID-ID) = один линтер `gid<slug>`
- `ruleguard/rules.go` — простые паттерн-правила (gocritic → ruleguard), слой 1
- `plugin.go` — регистрация всех анализаторов в plugin system
- `internal/pathseg` — матчинг слоёв по сегментам пути (`/domain/model`, `/dal/entity`, …)
- `internal/exclude` — разбор `settings.exclude` (`Метод` | `Тип.Метод`)
- `.golangci.yml` — эталонный конфиг: каждый линтер с `desc` и примерами настроек
- `RULES.md` — реестр правил со статусами; единственный источник истины по правилам

## Жёсткие требования

1. **Каждое правило обязано иметь eval.** Правило не считается готовым без
   `analysistest` + `testdata/src/...` с `// want`, покрывающего 4 класса кейсов:
   позитивный, негативный, граничный, неприменимость (шаблон — `rule_template.feature`).
2. Процесс добавления правила (конец RULES.md): строка реестра → `.feature`-спека →
   реализация → eval → включение в `.golangci.yml` → обновить статус в RULES.md.
3. UUID — только `github.com/gofrs/uuid` (мы сами форсируем это правилом GID-137).
4. Ошибки — только `github.com/pkg/errors` (GID-146).
5. eval-фикстуры в `testdata/` намеренно нарушают правила — не «чинить» их.

## Конвенции анализаторов

- Имя линтера: `gid<slug>` без дефисов (`gidnogetprefix`).
- Настройки через `settings` в `.golangci.yml`; точечные исключения через
  `//nolint:<линтер>`; централизованные — `settings.exclude` / `settings.tree` /
  `settings.tags` и т.п.
- Слой пакета определяется по сегментам пути через `internal/pathseg`,
  не по строковому `strings.Contains`.
- Диагностики и `description`/`Doc` формулируются **на английском** в формате
  `<problem>. Fix: <example>.` — каждое сообщение содержит валидный пример
  исправления. Соответственно `// want` в testdata пишутся на английском.
