# Реестр правил линтера

Источник правил: skill `go-styleguide` (внутренний стайлгайд gid.team).
Связанные документы: [PRD.md](PRD.md), [FINDINGS.md](FINDINGS.md) (аудит backend-go: исключения и упущенные правила), [FINDINGS_DDD.md](FINDINGS_DDD.md) (обратный аудит: конвенции из кода backend-go, кандидаты GID-218…223), [linter_framework.feature](linter_framework.feature), [rule_template.feature](rule_template.feature).

**Обязательное требование: каждое правило содержит eval.**
Правило не считается перенесённым, пока у него нет исполняемого eval:

- **go/analysis-правила** — `analysistest` + `testdata/src/...` с комментариями `// want`;
- **ruleguard-правила** — `analysistest` поверх анализатора `go-ruleguard` с нашим `rules.go` и тем же форматом `// want`;
- eval покрывает 4 класса кейсов: позитивный (нарушение ловится), негативный (чистый код проходит), граничный, неприменимость (см. чек-лист в `rule_template.feature`).

Статусы: ✅ готово · 🛠 в работе · 🔜 todo · ⛔ не переносим (остаётся на review)

---

## Слой 1: ruleguard (простые паттерны)

| ID | Слаг | Правило | Источник | Статус | Eval |
|---|---|---|---|---|---|
| GID-001 | no-time-now | Запрет `time.Now()` → `gdhelper.StdTime.Now()` | styleguide.md#временные-метки | ✅ | ✅ |
| GID-002 | no-uuid-empty-compare | Запрет `== uuid.UUID{}` / `!= uuid.UUID{}` → `IsNil()` | styleguide.md#идентификаторы | ✅ | ✅ |
| GID-003 | uuid-only-v7 | Генерация UUID только `uuid.Must(uuid.NewV7())`, запрет `NewV1/V3/V4/V5/V6` | styleguide.md#идентификаторы | ✅ | ✅ |
| GID-004 | allptr | `for range` по слайсу структур — через `gdhelper.AllPtr` (реализовано как go/analysis: нужны типы, чтобы отличить слайс структур от `[]*T`/`[]string`) | styleguide.md#итерация-по-слайсам | ✅ | ✅ |
| GID-005 | new-deref | `new(T)` для структур → `&T{}` | uber-guide + STYLEGUIDE_COVERAGE.md | ✅ | ✅ |
| GID-006 | yoda-conditions | Литерал слева в сравнении (`"foo" == x`) запрещён — переменная слева | google-decisions + STYLEGUIDE_COVERAGE.md | ✅ | ✅ |
| GID-007 | quote-verb | Ручное экранирование `\"%s\"` в format-строках → `%q` | google-decisions + STYLEGUIDE_COVERAGE.md | ✅ | ✅ |
| GID-008 | no-deepequal | `reflect.DeepEqual` запрещён — `cmp`/`require` в тестах, явное сравнение в коде | google-tests + STYLEGUIDE_COVERAGE.md | ✅ | ✅ |

## Слой 2: go/analysis (типы, имена, структура)

### Именование

| ID | Слаг | Правило | Источник | Статус | Eval |
|---|---|---|---|---|---|
| GID-101 | no-get-prefix | Запрет префикса `Get` в именах методов (искл.: сгенерированный код) | styleguide.md#именование-методов | ✅ | ✅ |
| GID-102 | no-batch-word | Запрет слова `Batch` в именах методов — множественное число вместо него | styleguide.md#именование-методов | ✅ | ✅ |
| GID-103 | receiver-naming | Ресивер — первая буква типа в нижнем регистре, две буквы для слайс-типов (искл.: `v` в validate-пакетах, `h` в handler-пакетах) | styleguide.md#именование-ресиверов | ✅ | ✅ |
| GID-104 | constructor-naming | Конструктор — `New<Entity>`, не голый `New` (искл.: composition root `internal/app/...`) | styleguide.md#структура-пакетов | ✅ | ✅ |
| GID-105 | converter-naming | Экспортируемые функции convert-пакетов именуются `<Dst><Type>From<Src>` (`ModelHelloOutFromEntity`) — линтер `gidconvnaming`. Покрывает и транспортные конвертеры (`ModelHelloFromGRPC`): scope включает `server` и `event` | converter.md#именование | ✅ | ✅ |
| GID-173 | iface-entity-prefix | Интерфейсы зависимостей — с префиксом сущности (`HelloRepository`); голые роли (`Repository`, `Service`, `Client`, `Connection`, `Producer`, `Consumer`, `Validator`, `Storage`, `Cache`) запрещены в service/usecase/repository/server/event; словарь настраивается `settings.names` (линтер `gidifacenaming`) | styleguide.md#интерфейсы + FINDINGS.md §2.1 | ✅ | ✅ |

### Сигнатуры

| ID | Слаг | Правило | Источник | Статус | Eval |
|---|---|---|---|---|---|
| GID-110 | ctx-first-param | `context.Context` — всегда первый параметр (линтер `gidparamorder`) | styleguide.md#сигнатуры-методов | ✅ | ✅ |
| GID-111 | input-ptr-output-value | Экспортируемые методы repo/service/usecase/handler: входные model/entity-структуры по указателю, выходные по значению. **Исключения**: `//nolint:gidinout` или `settings.exclude` (`Метод` \| `Тип.Метод`) | styleguide.md#сигнатуры-методов | ✅ | ✅ |
| GID-112 | create-update-no-return | Методы `Create*`/`Update*` в repo/service возвращают только `error`; данные получают отдельным запросом. **Исключения** (удобно сразу получить сущность): `//nolint:gidcreateupdate` или `settings.exclude` | styleguide.md#методы-создания + требование 2026-06-07 | ✅ | ✅ |
| GID-113 | opts-after-ctx | Параметр `opts` — первым после `ctx` (первым вообще, если ctx нет), не последним (линтер `gidparamorder`) | styleguide.md#options-паттерн | ✅ | ✅ |
| GID-114 | entity-method-naming | Методы repo/service содержат имя сущности: `Job`, `Jobs`, `JobsByStageID`, `CreateJob`; без суффикса `ByID` у одиночного получения (уточнения `By<Field>ID` разрешены), без префикса `List`. **Исключения** (Close, Ping, …): `//nolint:gidentitymethod` или `settings.exclude` (линтер `gidentitymethod`) | styleguide.md#именование-методов | ✅ | ✅ |

### Типы и поля

| ID | Слаг | Правило | Источник | Статус | Eval |
|---|---|---|---|---|---|
| GID-120 | no-ptr-uuid | Запрет `*uuid.UUID` в любой типовой позиции — проверка через `IsNil()` (линтер `gidnoptr`) | styleguide.md#nullable-поля | ✅ | ✅ |
| GID-121 | no-ptr-zero-checkable | В `/domain/model` поля структур без `*time.Time` и указателей на string-типы — zero-value выражает отсутствие (`IsZero()`, `len == 0`); `*bool` и вложенные структуры — допустимы (линтер `gidnoptr`) | styleguide.md#nullable-поля | ✅ | ✅ |
| GID-122 | entity-sql-null | В `/dal/entity` nullable-поля — `sql.NullString/NullTime/NullInt32/NullInt64` или `sql.Null[T]`, не указатели; фильтры (подпакет) не задеваются | styleguide.md#nullable-поля, entity.md | ✅ | ✅ |
| GID-123 | enum-string-based | Enum — именованный тип на основе `string`: в model/entity запрещены alias (`type X = string`), int-enum (именованный int-тип с ≥2 const) и группы из ≥2 нетипизированных string-const (линтер `gidenumbased`) | styleguide.md#enum + FINDINGS.md §1.7 | ✅ | ✅ |
| GID-124 | enum-string-method | Каждый enum (string-тип с const-значениями) реализует `String() string` | styleguide.md#enum | ✅ | ✅ |
| GID-125 | entity-db-tags | Экспортируемые поля entity-структур имеют тег маппинга на колонки БД — по умолчанию `db:`; список тегов настраивается `settings.tags` (например, `ch` у ClickHouse-библиотеки) | entity.md + требование 2026-06-07 | ✅ | ✅ |
| GID-126 | options-pattern | Тип настроек: постфикс `Options`, префикс сущности (голый `Options` запрещён вне app-слоя); package-level дефолты — `Default*` переменная. Композиция Options в app-слое (`Options` агрегирует `GRPCOptions`, `KafkaOptions`) — норма, не нарушение (FINDINGS.md §2.3) (линтер `gidoptsnaming`) | styleguide.md#options-паттерн | ✅ | ✅ |
| GID-168 | no-db-tags-in-model | В `/domain/**` запрещены теги маппинга на БД (`db:` по умолчанию, список — `settings.tags`) у полей структур — model чистый бизнес-объект, маппинг живёт в entity (линтер `gidmodeltags`) | model.md + FINDINGS.md §2.1 | ✅ | ✅ |
| GID-175 | in-transaction | Типы транзакций живут в `/domain/model` и называются `InTransactionFunc` / `InTransactionWithReturnFunc[T]`; в service/usecase анонимная tx-сигнатура запрещена — используется именованный тип из model; repo/service не оборачивают транзакцию методами — connection с tx-сигнатурой передаётся в конструктор напрямую (линтер `gidintransaction`) | требование 2026-06-07 | ✅ | ✅ |
| GID-210 | op-struct-fields | Операционные Create-структуры минимальны: в `/domain/model` `Create<X>` не содержит `ID`/`CreatedAt`/`UpdatedAt` (генерируются в service/convert); в `/dal/entity` `Create<X>` не содержит `UpdatedAt` — только поля INSERT (ID и CreatedAt в entity легитимны) (линтер `gidopstruct`) | model.md, entity.md | ✅ | ✅ |
| GID-211 | enum-location | Enum DAL-слоя (именованный string-тип с const-значениями) живёт только в `/dal/entity/enum` — отдельный файл на сущность; в model enum живёт прямо в model-слое (см. GID-132); alias (`type X = string`) — зона GID-123 (линтер `gidenumplace`) | entity.md | ✅ | ✅ |

### Структура файла и пакетов

| ID | Слаг | Правило | Источник | Статус | Eval |
|---|---|---|---|---|---|
| GID-130 | const-var-order | Порядок объявлений в файле: `import` → `const` → `var` → типы/функции; const всегда сверху, var под const | styleguide.md#порядок-объявлений + требование 2026-06-07 | ✅ | ✅ |
| GID-131 | no-upward-import | Дочерний пакет не импортирует родительский — общее выносится вниз, родитель импортирует детей (линтер `gidupwardimport`) | styleguide.md#направление-зависимостей | ✅ | ✅ |
| GID-132 | layer-imports | Направление слоёв: `/dal/**` не импортирует `/domain/**` (repo работает только с entity); `/domain/model` и `/domain/usecase` не импортируют `/dal/**` (usecase работает только с model, с DAL — через сервисы); `/domain/service` не импортирует `/dal/repository` (зависимость через интерфейс), импорт entity сервису разрешён. Слой матчится по сегментам пути: вложенные пакеты `/domain/model/*` (filter, enum, …) — полноправный model-слой, usecase может принимать/возвращать их типы | ARCHITECTURE.md + требование 2026-06-07 | ✅ | ✅ |
| GID-133 | methods-not-pkg-funcs | В service/usecase/repository приватные функции пакета запрещены — функция обязана быть методом структуры. Исключение: функция, используемая методами/конструкторами ≥2 сущностей пакета (общий хелпер). `convert/`/`build/` вне scope (отдельные пакеты) | styleguide.md#структура-пакетов + требование 2026-06-07 | ✅ | ✅ |
| GID-134 | interface-near-consumer | Интерфейсы живут там, где используются: интерфейс в полях/параметрах объявлен в том же пакете, использовать интерфейсы чужих пакетов нельзя. **Исключения**: интерфейсы библиотек (stdlib и внешние модули) и интерфейсы из `/domain/model` — для слоёв service/usecase (линтер `gidifaceplace`) | styleguide.md#интерфейсы + требование 2026-06-07 | ✅ | ✅ |
| GID-135 | convert-location | Функции-конвертеры (паттерн `<Dst>From<Src>`, кроме `*FromContext`) в слоях dal/domain/server/event живут в `convert/`-подпакете (линтер `gidconvnaming`) | converter.md | ✅ | ✅ |
| GID-136 | errors-new-static | `errors.New` (pkg/errors) только в package-level `var`, не в рантайме; `errors.Errorf` — не зона правила (линтер `giderrnew`; `err113` отклонён — конфликт с GID-146) | ARCHITECTURE.md | ✅ | ✅ |
| GID-137 | only-gofrs-uuid | Для UUID разрешена только `github.com/gofrs/uuid` — импорт google/uuid, satori/go.uuid, pborman/uuid, hashicorp/go-uuid, twinj/uuid запрещён | styleguide.md#идентификаторы | ✅ | ✅ |
| GID-138 | flat-layout | Репозитории и сервисы живут в корне `/dal/repository` и `/domain/service` — группирующие подпакеты (`/repository/redis`) запрещены; легитимные: `convert/`, `build/` (repo) и `convert/` (service) | требование 2026-06-07 | ✅ | ✅ |
| GID-148 | service-single | Сервис посвящён одной сущности и не зависит от другого сервиса (поле-структура из того же пакета `/domain/service`, кроме `*Options`) — оркестрация в usecase. Usecase может использовать несколько сервисов | требование 2026-06-07 | ✅ | ✅ |
| GID-170 | no-event-import | `/domain/**` и `/dal/**` не импортируют `/event/**` — event-слой зависит от domain/model и конвертирует model ↔ DTO, не наоборот (линтер `gidlayerimports`) | event.md, ARCHITECTURE.md + FINDINGS.md §2.1 | ✅ | ✅ |
| GID-171 | filter-location | Filter-структуры живут в своём месте слоя: в `/dal/**` — только `/dal/entity/filter`; в `/domain/**` — только model-слой (`/domain/model`, включая подпакеты). Только struct-типы; `FilterFunc` и `Filterable` не задеваются (линтер `gidfilterplace`) | model.md, entity.md + FINDINGS.md §2.1 | ✅ | ✅ |
| GID-172 | client-no-entity | `/client/**` не импортирует `/dal/**` — у клиента свои типы, он не знает о entity/repository (линтер `gidlayerimports`) | client.md + FINDINGS.md §2.1 | ✅ | ✅ |
| GID-174 | metric-prometheus-struct | Пакет метрик: путь `/metric` (не `metrics`), агрегатор `Prometheus` живёт в `prometheus.go` (только wiring) с методом `Register`, вызывающим `Register` каждой группы-поля; группы метрик — функциональные структуры, по одной на файл (линтер `gidmetricstruct`) | конвенция backend-go + требование 2026-06-07 | ✅ | ✅ |
| GID-194 | no-global-const | Константы объявляются там, где используются: package-level const вне `/domain/model/**` и `/dal/entity/**` запрещены — const, используемая одной функцией, объявляется внутри неё; неэкспортируемая, разделяемая ≥2 функциями пакета (или package-level объявлением, сигнатурой), остаётся package-level; экспортируемые — только в model/entity; iota-блок оценивается целиком. **Исключения**: `//nolint:gidconstscope` или `settings.exclude` (имена констант) (линтер `gidconstscope`) | требование 2026-06-07 | ✅ | ✅ |
| GID-195 | belongs-to-model | Приватная функция в `/domain/service`/`/domain/usecase` со строго одним параметром model-типа (`T`/`*T`, struct или enum), не зависящая от своего пакета, — поведение модели: оформляется публичным методом этого типа в model. Приватный метод, не использующий ресивер, — тот же случай. Непереносимые (используют ресивер, package-level символы, типы пакета в результатах), интерфейсы, generics, slice/variadic — не задеваются. Дополняет GID-133. **Исключения**: `//nolint:gidmodelmethod` или `settings.exclude` (`Функция` \| `Тип.Метод`) (линтер `gidmodelmethod`) | требование 2026-06-07 | ✅ | ✅ |
| GID-196 | chain-call-per-line | Цепочка вызовов из ≥`min-calls` звеньев (по умолчанию 2) — по одному вызову на строке, включая первый. Звено — вызов через селектор на результате другого вызова (включая промежуточные поля); конверсии — не звено; logrus-цепочки — зона GID-156; `_test.go` и сгенерированный код пропускаются (линтер `gidchainperline`) | перенос из «Частично проверяемые» + требование 2026-06-07 | ✅ | ✅ |
| GID-197 | interface-minimal | Интерфейс в service/usecase/repository/server/event содержит только методы, используемые в пакете-потребителе (вызов или метод-значение вне `_test.go`; GID-134 гарантирует, что потребитель в том же пакете). Embedded-интерфейсы не проверяются; FP-safe escape — интерфейс, чьё значение уходит под другим типом (any, assertion, constraint, неизвестный контекст), пропускается целиком. **Исключения**: `//nolint:gidifacemin` или `settings.exclude` (`Интерфейс` \| `Интерфейс.Метод`) (линтер `gidifacemin`) | перенос из «Частично проверяемые» + требование 2026-06-07 | ✅ | ✅ |
| GID-212 | build-signature | Экспортируемые функции `/dal/repository/build` возвращают `(string, []any, error)` (одиночный запрос) или `(*batch.Batch, error)` (batch-операции); импорт `github.com/Masterminds/squirrel` разрешён только в build-пакетах (линтер `gidbuildsig`) | repository.md, example_build.md | ✅ | ✅ |
| GID-215 | no-inline-entity-literal | В `/domain/**` (вне convert-пакетов) запрещён непустой composite literal entity-типа — конвертация model ↔ entity живёт в convert (`<Dst><Type>From<Src>`); пустой литерал (zero value) разрешён; флагается внешний литерал, вложенные не дублируются (линтер `gidinlineconv`) | service.md | ✅ | ✅ |
| GID-217 | ban-symbol | Настраиваемый бан символов библиотек (`settings.symbols`: pkg + name + msg; pkg матчится точно или по суффиксу сегментов); дефолт — `gdpostgres.TQuery` → «используй прямые методы conn: Select, ScanRow, NamedStruct, Transaction» (линтер `gidbansymbol`) | repository.md | ✅ | ✅ |
| GID-224 | transport-imports | Транспорт (`/server`, `/schedule`, `/validate`, `/event`) из слоёв сервиса видит только `/domain/model` (и `/validate`) — конкретные service/usecase инжектятся интерфейсами у потребителя; `/dal`, `/client`, `/metric`, `/app` и чужие транспорт-слои запрещены. Баны действуют внутри одного модуля (граница — сегмент `/internal/`). **Исключения**: `//nolint:gidlayerimports` или `settings.disable: [GID-224]` (линтер `gidlayerimports`) | решение 2026-06-07 (обсуждение depguard): строгая версия — в backend-go 25 импортов transport → service станут нарушениями | ✅ | ✅ |
| GID-225 | root-and-leaves | `/internal/app` (composition root) и транспорт-листья (`/server`, `/schedule`, `/validate`) никем не импортируются — wiring живёт в app, в транспорт никто не смотрит (линтер `gidlayerimports`) | решение 2026-06-07 | ✅ | ✅ |
| GID-226 | metric-standalone | `/metric` не импортирует слои сервиса (самостоятельный агрегатор Prometheus, ср. GID-174); `/domain` и `/dal` не импортируют `/metric` — метрики приходят интерфейсом, wiring в app (линтер `gidlayerimports`) | решение 2026-06-07; в backend-go 0 нарушений (кроме неканоничных event-collector/event-enricher) | ✅ | ✅ |
| GID-227 | model-pure | `/domain/model` не импортирует ни один слой сервиса — чистый словарь; подпакеты `/domain/model/*` — полноправный model-слой (дополняет GID-132/170) (линтер `gidlayerimports`) | решение 2026-06-07; в backend-go 0 нарушений | ✅ | ✅ |
| GID-228 | no-direct-client | `/domain/**` и `/dal/**` не импортируют `/client/**` — зависимость от клиента описывается интерфейсом в `/domain/model` (GID-134), конвертация model ↔ DTO клиента у потребителя/в app (линтер `gidlayerimports`) | решение 2026-06-07 | ✅ | ✅ |
| GID-229 | client-isolated | `/client/**` не импортирует слои сервиса, включая `/domain` целиком — у клиента свои типы (расширение GID-172). **Исключения**: `settings.disable: [GID-229]` на переходный период — в backend-go 11 импортов client → domain (линтер `gidlayerimports`) | решение 2026-06-07: строгое прочтение client.md | ✅ | ✅ |

### Обработка ошибок по слоям

| ID | Слаг | Правило | Источник | Статус | Eval |
|---|---|---|---|---|---|
| GID-176 | boundary-error-wrap | Ошибки, возникшие за пределами приложения (внешний вызов в `/client/**` и `/dal/repository`), всегда оборачиваются `errors.Wrap` — собирает стек и контекст. Внутри приложения (`/domain/**`) `errors.Wrap` для пришедшей нестатичной ошибки запрещён (стек уже собран на границе) — контекст добавляется `errors.WithMessage`, опционально. **Заменяет GID-140/141** (линтер `giderrwrap`) | repository.md, client.md + требование 2026-06-07 | ✅ | ✅ |
| GID-177 | static-error-withstack | Статичные ошибки (`ErrSome`, именованные error-типы `BigError`) при возврате всегда оборачиваются `errors.WithStack`; исключение — уже обёрнуты `errors.Wrap` (Wrap сам собирает стек). **Заменяет GID-142** (линтер `giderrwrap`) | service.md + требование 2026-06-07 | ✅ | ✅ |
| GID-143 | enum-convert-unhandled | Map-конвертация enum (в convert-пакетах: мапа с enum-ключом → именованный тип) обрабатывает отсутствующий ключ: comma-ok обязателен + в функции есть `gderror.NewUnhandledValueError` (линтер `gidenumconvert`) | converter.md#enum | ✅ | ✅ |
| GID-144 | domain-errors-in-model | Все domain-ошибки живут в `/domain/model`: в `/domain/**` вне model запрещены объявления error-переменных и конструкторы (`errors.New`, `fmt.Errorf`, `errors.Errorf`); `Wrap`/`WithStack`/`gderror.*` — разрешены | требование 2026-06-07 | ✅ | ✅ |
| GID-145 | dal-errors-in-entity | Все dal-ошибки живут в `/dal/entity` — симметрично GID-144 для `/dal/**` | требование 2026-06-07 | ✅ | ✅ |
| GID-169 | errors-in-error-file | Уточнение GID-144/145: в корневых пакетах `/domain/model` и `/dal/entity` error-переменные объявляются только в `error.go`/`errors.go`/`err.go` (список — `settings.files`; `err.go` — каноничное имя из entity.md) (линтер `giderrfile`) | model.md, entity.md + FINDINGS.md §2.1 | ✅ | ✅ |
| GID-146 | only-pkg-errors | Для работы с ошибками — только `github.com/pkg/errors`: std `errors.New`/`errors.Join`/`fmt.Errorf` запрещены везде; std `errors.Is/As/Unwrap` (проверка цепочки) — разрешены | требование 2026-06-07 | ✅ | ✅ |
| GID-147 | repo-returns-entity-errors | Репозиторий возвращает только ошибки из `/dal/entity`, обменивая ошибки подключения; без обмена — pass-through исходной. Детерминированное ядро покрыто GID-145 (не может создавать) + GID-132 (не может импортировать model); полная проверка потока возврата — review | требование 2026-06-07 | 🟡 ядро ✅ | — |
| GID-149 | service-returns-model-errors | service/usecase возвращают только ошибки из `/domain/model`. Ядро покрыто GID-144; полная проверка потока — review | требование 2026-06-07 | 🟡 ядро ✅ | — |
| GID-151 | service-model-api | API сервиса: экспортируемые методы в `/domain/service` принимают и возвращают model — entity-типы в параметрах/результатах запрещены (рекурсивно через указатели/слайсы/мапы); entity внутри тела (конвертация) — норма | требование 2026-06-07 | ✅ | ✅ |
| GID-157 | entity-group | Код сущности — единый блок: `type` → конструктор `New<Entity>` → методы, в файле объявления сущности; функции разных сущностей не перемешиваются | требование 2026-06-07 | ✅ | ✅ |
| GID-158 | dir-tree | Контроль дерева папок: для каждой папки из `settings.tree` разрешён только заданный перечень подпапок (дефолт — каноничная структура `internal/` из ARCHITECTURE.md: app, client, dal, domain, event, metric, server и вложенные). Чужая папка в `internal/` → подсказка «возможно, это service или usecase». Дерево редактируется, работает на любом уровне | требование 2026-06-07 | ✅ | ✅ |
| GID-159 | cache-in-repository | Кэш живёт в `/dal/repository`: кэширующий репозиторий оборачивает основной (прямой ссылкой, без интерфейса), вся магия с кэшом — в нём. Domain-слой про кэш не знает — импорт кэш-библиотек (redis, lru, ristretto, …) в `/domain/**` запрещён; список настраивается `settings.packages` | требование 2026-06-07 | ✅ | ✅ |
| GID-160 | grpc-via-repository | Service вызывает gRPC через repository: в `/domain/service` и `/domain/usecase` запрещён импорт `google.golang.org/grpc` и пакетов, которые сами импортируют grpc (pb-стабы). **Исключения** (иногда gRPC прямо в service): `//nolint:gidgrpcinservice` или `settings.exclude` | требование 2026-06-07 | ✅ | ✅ |
| GID-161 | no-panic | `panic` используется только в пакете `main` (bootstrap) — остальной код возвращает error. Это RULE-001 из PRD | PRD §5 + требование 2026-06-07 | ✅ | ✅ |

### Транспорт и валидация

| ID | Слаг | Правило | Источник | Статус | Eval |
|---|---|---|---|---|---|
| GID-162 | http-handler-own-errors | HTTP handler обрабатывает свои ошибки внутри себя: запрещены «супер-методы» (параметры `http.ResponseWriter` + `error` вместе) и хендлеры, возвращающие `error` | требование 2026-06-07 | ✅ | ✅ |
| GID-163 | http-data-response | HTTP handler строится на `github.com/raoptimus/data-response.go/v2` — чистый `func(w, r)` запрещён. **Исключения**: `//nolint:giddataresponse` или `settings.exclude` | требование 2026-06-07 | ✅ | ✅ |
| GID-164 | validator-go | Любые входящие данные (http, grpc, kafka) валидируются через `github.com/raoptimus/validator.go/v2`: validate-пакеты обязаны его использовать, сторонние валидаторы (go-playground, ozzo, govalidator) запрещены везде. **Исключения**: `settings.exclude` | требование 2026-06-07 | ✅ | ✅ |
| GID-165 | ctx-keys-in-model | Helper'ы, складывающие данные в context (`context.WithValue`), живут только в `/domain/model` — свой contextKey в middleware запрещён, иначе бизнес-слои зависят от middleware | требование 2026-06-07 | ✅ | ✅ |
| GID-166 | ctx-helper-naming | Форма ctx-helper'ов в model: кладёт в ctx → публичная `ContextWith<Name>`; достаёт → публичная `<Name>FromContext`; helper живёт в одном файле с сущностью `<Name>` (линтер `gidctxkeys`) | требование 2026-06-07 | ✅ | ✅ |
| GID-167 | context-key-type | Ключ контекста — публичный тип `ContextKey` (`type ContextKey string`), сырые/чужие типы ключей запрещены; все const-значения `ContextKey` — string в snake_case и находятся рядом с объявлением типа (в одном файле) (линтер `gidctxkeys`) | требование 2026-06-07 | ✅ | ✅ |
| GID-213 | validator-shape | Экспортируемый struct в validate-пакете (кроме имён с суффиксом `Options`) обязан иметь метод `Validate`: первый параметр `context.Context`, единственный результат `error`. **Исключения**: `//nolint:gidvalidatorshape` или `settings.exclude` (имена типов) (линтер `gidvalidatorshape`) | validator.md | ✅ | ✅ |
| GID-216 | event-ctor-deps | Конструктор kafka consumer принимает `*logrus.Logger`/`*logrus.Entry` (Entry собирается с полями broker/consumer — далее GID-154); конструктор producer logger не принимает — ошибки пробрасываются вызывающему коду. Подпакеты `validate`/`convert` не задеваются; schema-функции (`New<X>Schema`, тип чужого пакета) — не конструкторы. **Исключения**: `//nolint:gideventctor` или `settings.exclude` (имена конструкторов) (линтер `gideventctor`) | event.md + FINDINGS.md §2.4 | ✅ | ✅ |

### Options и логирование (logrus)

| ID | Слаг | Правило | Источник | Статус | Eval |
|---|---|---|---|---|---|
| GID-152 | opts-style | opts в параметрах передаётся указателем (`*XxxOptions`); opts в теле сущности встраивается (embedded), а не хранится именованным полем | требование 2026-06-07 | ✅ | ✅ |
| GID-153 | logger-after-opts | logger всегда идёт после opts сущности, если оба существуют (линтер `gidparamorder`) | требование 2026-06-07 | ✅ | ✅ |
| GID-154 | logger-withfield-ctor | Если сущность содержит logger (logrus), конструктор `New<Entity>` обязан вызвать `WithField(<entity>, <name>)` | требование 2026-06-07 | ✅ | ✅ |
| GID-155 | log-withcontext-witherror | Лог-вызов в функции с `ctx` обязан содержать `WithContext`; лог уровня `Error*` обязан содержать `WithError`. «WithError если есть error в области видимости» в общем виде — review | требование 2026-06-07 | ✅ | ✅ |
| GID-156 | log-chain-multiline | Цепочка logrus из ≥2 вызовов — каждый вызов на своей строке, включая первый; одиночный вызов inline допустим | требование 2026-06-07 | ✅ | ✅ |
| GID-214 | logger-singleton | `logrus.New()`/`logrus.StandardLogger()` вызываются только в composition root (пакет `main`, `internal/app`) — остальной код получает готовый `*logrus.Entry` через конструктор (GID-153/154); `_test.go` не задеваются (линтер `gidloggernew`) | libs.md | ✅ | ✅ |

### Перенесено из Uber / Google best practices (STYLEGUIDE_COVERAGE.md)

| ID | Слаг | Правило | Источник | Статус | Eval |
|---|---|---|---|---|---|
| GID-178 | no-embed-mutex | Запрет встраивания `sync.Mutex`/`sync.RWMutex` в структуры — мьютекс хранится именованным неэкспортируемым полем | uber: zero-value mutex | ✅ | ✅ |
| GID-179 | chan-buffer-size | Буфер канала только 0 или 1: `make(chan T, N)` с литералом N>1 запрещён. **Исключения**: `//nolint:gidchanbuf` | uber: channel size | ✅ | ✅ |
| GID-180 | init-clean | В `init()` запрещены запуск goroutine и I/O (os/net/db-вызовы) — init детерминированный | uber: avoid init | ✅ | ✅ |
| GID-181 | exit-once | `os.Exit`/`log.Fatal*` вызываются не более одного раза и только в `main()` (дополняет GID-161) | uber: exit once | ✅ | ✅ |
| GID-182 | bytes-in-loop | Конверсия строкового литерала/константы в `[]byte`/`[]rune` внутри цикла — вынести из цикла | uber: perf | ✅ | ✅ |
| GID-183 | map-capacity-hint | `make(map...)` без capacity + заполнение в `range`-цикле по коллекции с известным `len` → указать хинт | uber: perf (prealloc не умеет map) | ✅ | ✅ |
| GID-184 | no-failed-to | Бан префиксов `failed to`/`failed`/`unable to` в сообщениях `Wrap`/`WithMessage`/`Errorf` — сообщение описывает операцию, не факт провала | uber: error wrapping | ✅ | ✅ |
| GID-185 | nil-slice-style | `return []T{}` → `return nil`; `s := []T{}` → `var s []T` (zero-value слайс валиден) | uber/google: nil slices | ✅ | ✅ |
| GID-186 | format-string-const | Format-строка printf-функций — литерал или `const`, не переменная | uber: format strings | ✅ | ✅ |
| GID-187 | no-util-package | Бан имён пакетов `util`/`utils`/`common`/`helper`/`helpers`/`shared`/`misc`; список — `settings.names` | google: util packages | ✅ | ✅ |
| GID-188 | no-custom-context | Запрет кастомных context-типов: в позиции ctx-параметра и в `interface`-embedding только `context.Context` | google: custom contexts (no exceptions) | ✅ | ✅ |
| GID-189 | chan-direction | Параметры-каналы в сигнатурах — с направлением (`<-chan`/`chan<-`), двунаправленный параметр запрещён | google: channel direction | ✅ | ✅ |
| GID-190 | error-last | `error` — последний возвращаемый параметр; конкретные error-типы в результатах запрещены (typed-nil ловушка) — возвращается интерфейс `error` | google: errors | ✅ | ✅ |
| GID-191 | subtest-naming | Имена subtest в `t.Run` — без пробелов и слешей (литералы) | google: subtest names | ✅ | ✅ |
| GID-192 | flag-in-main | `flag.*`/регистрация флагов только в пакете `main`; имя флага snake_case, переменная camelCase | google: flags | ✅ | ✅ |
| GID-193 | no-pkg-stutter | Экспортируемый символ не повторяет имя пакета (`widget.NewWidget`, `widget.WidgetOptions`). **Исключение**: конструкторы `New<Entity>` (GID-104 главнее) | google: repetition | ✅ | ✅ |

Отброшено как противоречащее нашему стайлгайду: `_`-префикс приватных глобалов (Uber), enum start at one (у нас string-enum — GID-123), `var _ Iface` compliance-assertions (judgment), go.uber.org/atomic.

## Слой 3: покрыто стандартными линтерами golangci-lint

| ID | Правило | Линтер | Статус |
|---|---|---|---|
| GID-201 | Ширина строки ≤ 120 (осознанное отклонение от Google) | `lll` (`line-length: 120`) | ✅ конфиг |
| GID-202 | Запрет `_ = err`; обязательный comma-ok у type assertion | `errcheck` (`check-blank`, `check-type-assertions`) | ✅ конфиг |
| GID-203 | Форматирование goimports | formatters: `goimports` | ✅ конфиг |
| GID-204 | Динамические ошибки в рантайме | `err113` включён базой consent-api, его dynamic-errors-чек погашен exclusion'ом (конфликт с GID-146 pkg/errors); закрыто GID-136/144/145/146 | ✅ конфиг |
| GID-205 | Корректность: copylocks, composites, printf, lostcancel, SA-чеки, unused, ineffassign | стандартная пятёрка: `govet`, `staticcheck`, `unused`, `ineffassign` (+`errcheck`) | ✅ конфиг |
| GID-206 | Стиль staticcheck: mixed caps/initialisms (ST1003), package comment (ST1000), консистентный ресивер (ST1016), error strings (ST1005), error naming (ST1012) | `staticcheck` `checks: [all]` | ✅ конфиг |
| GID-207 | Doc-comments экспортируемых, indent-error-flow, early-return, superfluous-else, deep-exit, dot/blank imports, `any` вместо `interface{}` | `revive` (точечные rules) | ✅ конфиг |
| GID-208 | Builtin-шейдинг, запрет init, naked return, slice capacity, strconv vs fmt, теги marshaled-структур, вложенность, когнитивная сложность, ctx в структуре, размер интерфейса, return concrete types, группировка объявлений, proto-алиасы | `predeclared`, `gochecknoinits`, `nakedret`, `prealloc`, `perfsprint`, `musttag`, `nestif`, `gocognit`, `containedctx`, `interfacebloat`, `ireturn`, `grouper`, `importas` | ✅ конфиг |
| GID-209 | Тесты: `t.Helper()`, `_test`-пакеты, корректность параллельных тестов | `thelper`, `testpackage`, `tparallel`, `paralleltest` (`ignore-missing`) | ✅ конфиг |

С переходом на базу consent-api (2026-06-07): `errorlint`, `err113`, `forcetypeassert` включены
базой (dynamic-errors-чек err113 погашен exclusion'ом — конфликт с GID-146; forcetypeassert
дублирует `errcheck.check-type-assertions` — вреда нет). `depguard` включён для
модуленезависимых банов библиотек (uuid-форки — GID-137 на уровне импорта); запреты слоёв
в depguard не выражаются (deny — префиксный матчинг полного import-пути, требует хардкода
module path в каждом сервисе) — их закрывает `gidlayerimports` (GID-132/170/172/224…229).
Не включаем: `gochecknoglobals` (ломается об `Default*Options` и `var ErrX`), `wrapcheck`
(заменён собственным `giderrwrap` — GID-176/177).

## Не переносим (остаётся на code review) ⛔

Из «частично проверяемых» реализованы и перенесены в реестр: chain-call-per-line → GID-196,
interface-minimal → GID-197 (2026-06-07); раздел закрыт — остальные эвристики решением
2026-06-07 признаны непереносимыми (FP-риск неприемлем) и перечислены ниже.

Качество имён · глагольность имён операций (verb-first-naming: словарь глаголов даёт высокий FP — детерминированной формулировки нет) · уместность абстракций · полнота обработки ошибок · сложность методов · читаемость · корректность бизнес-логики · полнота логирования/метрик · корректность SQL в build-функциях · правильность FSM-переходов · возвращаемый тип с полями, которые функция не может заполнить (styleguide.md#возвращаемые-типы) · стиль SQL-параметров (`@id` в inline-SQL vs `$1` в build-функциях — требует анализа строк запросов) · порядок полей entity-структур (ID → бизнес-поля → nullable → timestamps; высокий FP-риск) · new-prefix-factories (неэкспортируемые фабрики с префиксом `new`: эвристика «что такое фабрика» даёт неприемлемый FP) · ptr-at-callsite (`&` в месте вызова, не `&Type{}` заранее: usage-анализ даёт спорные диагностики).

---

## Процесс добавления правила

1. Завести строку в этом реестре (ID, слаг, источник).
2. Спецификация — `.feature` по шаблону `rule_template.feature` (4 класса кейсов).
3. Реализация: ruleguard-функция в `ruleguard/rules.go` **или** анализатор в `analyzers/<slug>/`.
4. **Eval обязателен**: testdata с `// want`, все 4 класса кейсов, `go test ./...` зелёный.
5. Включить в `.golangci.yml`, обновить статус здесь.
