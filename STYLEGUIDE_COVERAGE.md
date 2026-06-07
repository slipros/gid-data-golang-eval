# Покрытие best-practices стайлгайдов линтерами golangci-lint v2

Источники: [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md),
[Google Go Style Guide](https://google.github.io/styleguide/go/guide),
[Google Go Best Practices](https://google.github.io/styleguide/go/best-practices).

Бакеты:
- **DEFAULT** — ловится дефолтным набором golangci-lint v2 (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`);
- **OPT-IN** — есть готовый линтер/форматтер, нужно включить и настроить;
- **CUSTOM** — готового линтера нет, нужен наш go/analysis или ruleguard;
- **REVIEW** — не автоматизируется, остаётся на code review.

> ⚠️ **Нюанс staticcheck в v2:** при включённом линтере `staticcheck` по умолчанию
> работают `all` **минус** `ST1000` (package comment), `ST1003` (mixed caps /
> initialisms), `ST1016` (консистентный ресивер), `ST1020-22` (doc-comments).
> То есть SA (баги), S (simple), QF и большинство ST-проверок (включая ST1005
> error strings, ST1012 error naming, ST1006 запрет this/self) — активны.
> Исключённые ST включаются через `settings.staticcheck.checks`.

> ⚠️ **Наш эталонный `.golangci.yml` сейчас `default: none`** — из стандартной
> пятёрки включён только `errcheck`. `govet`, `staticcheck`, `ineffassign`,
> `unused` не работают. Это первый кандидат на исправление.

---

## 1. Что закрывает дефолтная пятёрка golangci-lint v2

| Правило из гайдов | Линтер/чек |
|---|---|
| Не игнорировать ошибки, `_ = err` (Google: handle errors) | `errcheck` (+ `check-blank: true` — уже GID-202) |
| Не копировать мьютексы/locks по значению (Uber: zero-value mutex; Google: copying) | `govet/copylocks` |
| Имена полей в composite literals чужих пакетов (Uber/Google) | `govet/composites` |
| Printf: соответствие формата аргументам, имя `...f` (Uber) | `govet/printf` |
| Утечка cancel у контекста (Google: goroutine lifetimes, частично) | `govet/lostcancel` |
| Захват loop-переменной (Uber: parallel tests; до Go 1.22) | `govet/loopclosure` |
| Ключ контекста — не базовый тип (Google: context keys) | `staticcheck SA1029` |
| Typed-nil в interface-сравнении (Google: return error interface, частично) | `staticcheck SA4023` |
| `fmt.Errorf` без форматирования → `errors.New` (Uber) | `staticcheck S1028` |
| Error strings: lowercase, без точки (Uber/Google) | `staticcheck ST1005` |
| Нейминг ошибок `ErrX`/`errX` (Uber) | `staticcheck ST1012` |
| Ресивер не `this`/`self`/`me` (Google) | `staticcheck ST1006` |
| Redundant `break`, избыточные nil-чеки и пр. simple-кейсы | `staticcheck S1023`, `S1009`, `S1021`, … |
| Неиспользуемый код / бесполезные присваивания | `unused`, `ineffassign` |

**Вывод:** дефолт закрывает пласт «корректность + базовая гигиена ошибок», но
почти ничего из структурно-стилевого слоя гайдов.

## 2. OPT-IN: готовые линтеры, которые стоит обсудить к включению

### Высокая ценность, конфликтов с нашим стайлгайдом нет

| Линтер | Что закрывает из гайдов | Примечание |
|---|---|---|
| `govet` + `staticcheck` + `unused` + `ineffassign` | весь раздел 1 | сейчас выключены из-за `default: none` |
| `revive` (точечные правила) | `indent-error-flow`, `early-return`, `superfluous-else` (Uber: nesting/else), `deep-exit` (Uber: exit in main; Google: log.Fatal), `dot-imports`, `blank-imports` (Google), `use-any`, `exported` (doc-comments, Google) | включать только нужные rules |
| `staticcheck ST1003, ST1016, ST1000` | mixed caps, initialisms (URL не Url), консистентный ресивер, package comment (Google) | добавить в `settings.staticcheck.checks` |
| `gocritic` (шире, чем ruleguard) | десятки мелких правил Uber: `ifElseChain`, `exitAfterDefer`, `ptrToRefParam` (указатель на интерфейс) и др. | каркас уже подключён ради ruleguard |
| `predeclared` | Uber: не затенять builtin-имена | — |
| `gochecknoinits` | Uber: avoid `init()` | bootstrap в `internal/app` — исключения |
| `nakedret` | Google: naked return только в коротких функциях | — |
| `forcetypeassert` *или* `errcheck.check-type-assertions: true` | Uber: comma-ok при type assertion | у errcheck опция выключена по умолчанию |
| `prealloc` | Uber: slice capacity hints | map-хинты не покрывает |
| `perfsprint` | Uber: strconv vs fmt, Sprintf→конкатенация (Google: string concatenation) | — |
| `musttag` | Uber: field tags в marshaled structs | дополняет наши GID-125/168 |
| `nestif`, `gocognit`/`gocyclo` | Uber: сложность table-subtests, nesting | пороги обсудить |
| `importas` | Google: алиасы proto/`pb`-импортов | — |
| `thelper`, `testpackage`, `paralleltest`, `tparallel` | Google tests: `t.Helper()`, `_test`-пакеты, параллельность | согласовать со skill go-testing |
| `containedctx` | Google: не хранить Context в структуре | — |
| `interfacebloat` | Google: маленькие интерфейсы | порог методов |
| `ireturn` | Google: accept interfaces, return concrete types | настроить allow-list (error, generics) |
| `grouper` | Uber: группировка const/var/import в блоки | дополняет GID-130 |
| `wrapcheck` | **наше правило «ошибки извне всегда Wrap»** (≈GID-140/141) | проверяет, что err из чужого пакета обёрнут; настраивается под pkg/errors — кандидат вместо/в основу custom-правила |

### Конфликтуют с нашим стайлгайдом — включать нельзя или с оговорками

| Линтер | Конфликт |
|---|---|
| `errorlint` (`%w`, `errors.Is`) | гайды строятся на std-wrapping `%w`; у нас GID-146 — только `pkg/errors`, `fmt.Errorf` запрещён. Полезна только часть `comparison`/`asserts` (`errors.Is`/`As` вместо `==`) — std `errors.Is/As` у нас разрешены |
| `gochecknoglobals` | у нас package-level `var Default*Options` (GID-126) и `var ErrX` — легитимны. Нужны исключения или не включать |
| `err113` | то же: толкает к std errors; пересечение с GID-136 решить в пользу нашего правила |
| `depguard` | мог бы заменить GID-137/146 (бан uuid-форков, testify по Google) — но наши custom-линтеры дают лучшие сообщения; Google запрещает assert-библиотеки, у нас testify (require/assert) — осознанное отклонение |
| `lll` | Google прямо называет line-length «invalid local style»; наш GID-201 (120) — осознанное отклонение, оставляем |

## 3. CUSTOM: кандидаты на новые GID-правила

Детерминируемые (AST/types), готового линтера нет, с нашими правилами не пересекаются:

| Кандидат | Источник | Комментарий |
|---|---|---|
| Запрет embed `sync.Mutex`/`RWMutex` в структуру | Uber | тривиальный AST-чек |
| Буфер канала только 0 или 1 (`make(chan T, N)`, N>1 — диагностика) | Uber | литералы; настройка исключений |
| Запрет goroutine в `init()` / I/O в `init()` | Uber | дополняет gochecknoinits, если init разрешим в app |
| `os.Exit`/`log.Fatal` не более одного раза в `main` | Uber | счётчик вызовов |
| `[]byte("literal")` внутри цикла → вынести | Uber perf | — |
| Map capacity hint (`make(map, n)` при известном size) | Uber perf | prealloc не умеет |
| Бан `"failed to ..."` в сообщениях ошибок | Uber | строковый литерал в Wrap/WithMessage |
| `return []T{}` → `return nil`; `var s []T` vs `s := []T{}` | Uber/Google | — |
| `new(T)` → `&T{}`; `T{}` → `var x T` для zero-value | Uber | ruleguard-уровень |
| Format string — `const`/литерал, не переменная | Uber | дополняет govet/printf |
| Бан имён пакетов `util`/`common`/`helper`/`shared` | Google | расширение GID-158 (dirtree) или отдельное |
| Запрет кастомных context-типов (не `context.Context` в позиции ctx) | Google | «no exceptions» у Google |
| Направление каналов в сигнатурах (`<-chan`/`chan<-`) | Google | types-чек параметров |
| `error` — последний возвращаемый параметр; не возвращать конкретный error-тип | Google | typed-nil ловушка |
| Yoda conditions (`"foo" == x`) | Google | ruleguard |
| `%q` вместо ручных `\"%s\"` | Google | ruleguard |
| `reflect.DeepEqual` в тестах → cmp/require | Google | ruleguard |
| Имена subtest в `t.Run` без пробелов/слешей | Google | литералы |
| `flag.*` только в `package main`; имя флага snake_case | Google | если появятся CLI |
| Символ не дублирует имя пакета (`widget.NewWidget`) | Google | дополняет GID-104 |

Отброшено как противоречащее нашему стайлгайду: `_`-префикс приватных глобалов
(Uber-специфика), enum start at one (у нас string-enum, GID-123), верификация
interface compliance через `var _` (judgment), go.uber.org/atomic (своя экосистема).

## 4. REVIEW: не автоматизируется

Clarity/Simplicity/Consistency, выбор типа ошибки по матрице, in-band errors,
дизайн интерфейсов («потребитель определяет» — у нас частично закрыто GID-134/173),
goroutine lifecycle, time-семантика, functional options vs option struct,
got-before-want в тестах, полнота документации, global state litmus tests.

---

## Сводка

| | Uber | Google guide+BP | Итого уникальных |
|---|---|---|---|
| DEFAULT (после включения пятёрки) | ~7 | ~9 | ~14 |
| OPT-IN | ~18 | ~25 | ~25 линтеров/правил |
| CUSTOM детерминируемый | ~25 | ~15 | ~20 кандидатов |
| REVIEW | ~12 | ~30 | — |

Приоритет обсуждения:
1. Включить стандартную пятёрку в эталонный конфиг (`govet`, `staticcheck`, `unused`, `ineffassign`).
2. Добрать дешёвые OPT-IN: `revive` (точечно), ST1003/ST1016/ST1000, `predeclared`, `nakedret`, `forcetypeassert`, `prealloc`, `perfsprint`, `musttag`.
3. `wrapcheck` — как основа/замена будущих GID-140/141 (ошибки извне → Wrap).
4. Из CUSTOM-кандидатов выбрать первую волну (предлагаю: embed mutex, channel size, exit-once, util-пакеты, custom context, error last param, DeepEqual в тестах).
