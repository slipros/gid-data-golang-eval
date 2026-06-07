# Аудит: исключения в backend-go и упущенные правила

Дата: 2026-06-07.
Источники: [RULES.md](RULES.md), skill `go-styleguide` (полный gap-анализ доков),
скан 20 сервисов `/mnt/w/GPM-Data/UDMP/backend-go`.
Обратное направление (конвенции из кода, не попавшие в стайлгайд) — [FINDINGS_DDD.md](FINDINGS_DDD.md).

Документ фиксирует: (1) места, где реальный код нарушает уже реализованные правила —
кандидаты в `settings.exclude` или сигнал к смягчению правила; (2) правила, упущенные
в реестре, и решение по каждому; (3) открытые вопросы.

---

## 1. Исключения — нарушения реализованных (✅) правил

### 1.1 GID-001 `no-time-now` — ~15 мест, три кластера

| Кластер | Файлы |
|---|---|
| **Замер метрик** (`defer calculateMetrics(..., time.Now())`) | `file-storage/internal/client/minio.go:101,123,172,225,251` · `profile-targeting-api/internal/client/http/metrics.go:8` · `file-storage/internal/app/file-storage/application.go:212` |
| **`nowFunc` в валидаторах** (max-дата = «сейчас») | `consent-api/pkg/consent/v{1,2}/server/http/handler/validate/statuses.go:35` · `event-api/internal/server/http/router/handler/validate/consent_event_v{1,2}.go:90` · `gid-sso-consent-consumer/internal/event/consumer/kafka/validate/consent_event_v1.go:80,215`, `consent_event_v2.go:96,484` |
| **Шедулеры** | `consent-api/internal/schedule/schedule.go:20` · `data-marketplace/internal/schedule/schedule.go:24` · `data-marketplace/pkg/marketplace/job/materialization/job.go:144` |

**Вопрос**: смягчить правило для замеров длительности в metric-коде — или всё в `settings.exclude`.

### 1.2 GID-112 `create-update-no-return` — ~12 мест

| Сервис | Файл | Метод |
|---|---|---|
| consent-api | `internal/domain/service/attribute_group.go:119` | `Create() (uuid.UUID, error)` |
| consent-api | `internal/domain/service/subject_group.go:119` | `Create() (uuid.UUID, error)` |
| consent-api | `internal/domain/service/webhook.go:66` | `Create() (uuid.UUID, error)` |
| data-marketplace | `pkg/datalab/domain/service/showcase.go:100` | `Create() (uuid.UUID, error)` |
| data-marketplace | `pkg/marketplace/domain/service/tags.go:34` | `Create() (uuid.UUID, error)` |
| data-marketplace | `pkg/marketplace/domain/service/dataset.go:264` | `CreateDataset() (uuid.UUID, error)` |
| data-marketplace | `pkg/marketplace/dal/repository/dataset.go:175` | `PutTable() (uuid.UUID, error)` |
| data-marketplace | `pkg/marketplace/dal/repository/tag.go:36` | `Create() (*omd.CreatedTag, error)` |
| organization-ticket | `internal/domain/service/organization_ticket.go:194` | `CreateAdvertisingCampaignTicket() (uint64, error)` |
| organization-ticket | `internal/dal/repository/organization_ticket.go:145` | `Create() (uint64, error)` |
| user-api | `internal/service/field_activity/service.go:81` | `Create() (uuid.UUID, error)` |
| user-api | `internal/service/organization/service.go:36` | `Create() (uuid.UUID, error)` |
| user-api | `internal/service/group_organization/service.go:34` | `Create() (uuid.UUID, error)` |

Готовый список для `settings.exclude` / `//nolint:gidcreateupdate` — либо техдолг к починке.

### 1.3 GID-144/145 `errors-in-model/entity` — самое массовое (50+ файлов)

`errors.go` с `errors.New` живёт в `/domain/service/` и `/dal/repository/` почти везде:

- **lk-api** — 30+ файлов: `internal/domain/service/errors.go`, `internal/validate/errors.go`, `pkg/{account,member,consent,organization,…}/domain/service/errors.go`, `pkg/*/dal/repository/errors.go`
- **organization-portfolio** — `internal/domain/service/errors.go`, `internal/dal/repository/errors.go`
- **organization-ticket** — `internal/domain/service/errors.go:6-12`
- **file-storage** — `internal/service/errors.go` (26 ошибок; сама папка нарушает GID-158), `internal/client/errors.go`, `internal/dal/repository/errors.go`
- **event-packer** — `internal/domain/service/errors.go:9`
- **gid-sso-consent-consumer** — `internal/event/consumer/kafka/validate/errors.go`

**Ключевой вопрос**: правило = целевое состояние (это техдолг) или норму команды надо пересмотреть.

### 1.4 GID-160 `grpc-via-repository` — ложноположительный класс

В `/domain/service` импортируются только `google.golang.org/grpc/codes` и `…/status`
для **маппинга кодов ошибок**, не для транспорта:
`gid-sso-consent-consumer/internal/domain/service/user_v2.go:10-11`,
`lk-api/internal/domain/service/{file,file_upload,organizations}.go`.

**Предложение**: разрешить `codes`/`status` в правиле — иначе шквал nolint.

### 1.5 GID-158 `dir-tree` — повторяющиеся «нестандартные» папки

| Папка | Сервисы | Оценка |
|---|---|---|
| `schedule/` | consent-api, data-marketplace | кандидат в дефолтный allowlist |
| `validate/` | file-storage, lk-api, consent-api | обсудить (валидация общего назначения) |
| `statement/`, `batch/`, `collector/`, `consumer/` | event-collector | специализированный пайплайн |
| `enricher/` | event-enricher | специализированный пайплайн |
| `job/`, `producer/`, `service/` | file-storage | `service/` — нарушение (должно быть `domain/service`) |
| `metrics/` (вместо `metric`), `testharness/` | lk-api | `metrics` → переименовать |

### 1.6 Кандидаты на исключение целыми модулями

- **tools-cli** — CLI-утилита (urfave/cli) без слоёв: исключить из GID-110/111/130-х/132/151/158.
- **user-api** — легаси-структура (`/internal/service/{account,member,…}` вместо `/domain/service`); нарушает GID-132, GID-148 (`member` → `organizationService`, `auditLog`, `providerService`; `account` → `memberService` + репозиторий напрямую), GID-158.
- **test/**, **example/** (event-collector: `example/http-request-producer/producer.go:8` импортирует `github.com/google/uuid` → GID-137).

### 1.7 Прочее

- **GID-123, реальный кейс**: event-collector `internal/domain/model/enum/{consent_event_type,consent_event_v2_type,device_type}.go` — `type X = string` (alias вместо именованного типа). Берём как позитивный кейс eval'а GID-123.
- **GID-003 в тестах**: `uuid.Must(uuid.NewV4())` массово в `_test.go` (consent-webhook-trigger — 30+, data-reports). Вопрос: применять ли правило к тестам.

---

## 2. Упущенные правила

### 2.1 Принято в работу (реализуются в этом репо)

| ID | Слаг | Правило | Источник |
|---|---|---|---|
| GID-123 | enum-string-based | Реализация существующего 🔜: именованный string-тип, запрет alias (`type X = string`) и int-enum, запрет групп нетипизированных string-const в model/entity | styleguide.md#enum + кейс event-collector |
| GID-168 | no-db-tags-in-model | В `/domain/**` запрещены `db:`-теги у полей структур — маппинг на БД живёт в entity | model.md |
| GID-169 | errors-in-error-file | Уточнение GID-144/145: error-переменные в `/domain/model` и `/dal/entity` живут в файле `error.go`/`errors.go` | model.md, entity.md |
| GID-170 | no-event-import | `/domain/**` и `/dal/**` не импортируют `/event/**` — event-слой конвертирует model ↔ DTO, не наоборот | event.md, ARCHITECTURE.md |
| GID-171 | filter-location | Filter-структуры: в `/dal/**` — только `/dal/entity/filter`; в `/domain/**` — только model-слой | model.md, entity.md |
| GID-172 | client-no-entity | `/client/**` не импортирует `/dal/**` — у клиента свои типы | client.md |
| GID-173 | iface-entity-prefix | Интерфейсы зависимостей — с префиксом сущности (`HelloRepository`), голые роли (`Repository`, `Service`, …) запрещены | styleguide.md#интерфейсы |
| GID-174 | metric-prometheus-struct | Пакет `/metric`: тип `Prometheus` (агрегатор метрик по протоколам) с методом `Register`; пакет называется `metric` | конвенция всех сервисов |
| GID-175 | in-transaction | Типы транзакций (`InTransactionFunc`, `InTransactionWithReturnFunc[T]`) живут в `/domain/model`; service/usecase используют именованный тип (connection с tx-сигнатурой инжектится конструктором); repo/service не объявляют tx-методы | требование 2026-06-07 |

Каноническая форма GID-175 (живёт в `/domain/model`):

```go
type InTransactionFunc func(ctx context.Context, fn func(ctx context.Context) error) error

type InTransactionWithReturnFunc[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)

// NewInTransactionWithReturnFunc wraps InTransactionFunc to support a return value.
func NewInTransactionWithReturnFunc[T any](tx InTransactionFunc) InTransactionWithReturnFunc[T] { … }
```

### 2.2 Уже покрыто существующими правилами (реализация не нужна)

- **transport-converter-naming** (`Model<T>FromGRPC` / `GRPC<T>FromModel`) — `gidconvnaming` (GID-105/135) уже действует в слоях `server` и `event`, паттерн `<Dst><Type>From<Src>` покрывает транспортные имена. Не покрыт только нейминг *приватных* enum-хелперов (`grpc<X>FromEnum`) — низкая ценность, отложено.
- **dal-enum-string** — GID-124 (`gidenumstring`) не ограничен слоем и уже требует `String()` у любых string-enum, включая `/dal/entity/enum`; «именованный тип, не alias» закрывается реализацией GID-123.

### 2.3 Найденное противоречие — nested Options

Из доков (app.md) выводится «Options не содержат вложенных `*Options`», но реальный код
повсеместно делает наоборот: `Options` в `internal/app/api/options.go` агрегирует
`GRPCOptions`, `KafkaOptions`, `TraceOptions` (consent-api, event-api, event-collector,
organization-ticket, …). Решение: кандидат отклонён; при доработке GID-126 учесть, что
композиция Options в app-слое — норма.

### 2.4 Отложено (средняя ценность, не взяты сейчас)

| Слаг | Правило | Почему отложено |
|---|---|---|
| ctor-deps-signature | Часть про event-слой (consumer принимает logger, producer — нет) **реализована как GID-216** (2026-06-07). Остаток: Client обязан принимать Metrics (client.md) — отложен до уточнения спеки client-слоя | узкое, детерминированное, но мало нарушений |
| slice-type-required | Для основной сущности обязателен слайс-тип (`Jobs []Job`) | нет надёжного признака «основной» сущности — FP |
| logger-nil-default | Конструктор с logger: nil → `logrus.StandardLogger()` | в доках не зафиксировано, только в коде 3 сервисов |
| mocks-location | Моки в `mock*/` подпакетах | конвенция из кода, в стайлгайде нет |
| private-enum-converter-naming | `grpc<X>FromEnum`, `avro<X>FromDTO` для приватных конвертеров | низкая ценность |

### 2.5 Остаётся на review (статически не проверить)

Validator-паттерны (`When` vs callback, `NewNested`+`NewEach`, internal vs public валидация,
proto3-enum через `NewInRange`, `NewTime` только для string) · «конвертер — чистая функция» ·
«handler не содержит бизнес-логики» · «usecase зависит от сервисов, не от repo» (частично
закрыто GID-132) · FSM-карта переходов + `CanTransitionTo` · graceful shutdown в main ·
генерация ID/CreatedAt в конвертере · операционные entity-структуры с минимальным набором полей.

---

## 3. Открытые вопросы (требуют решения владельца стайлгайда)

1. **GID-144/145 vs реальность** — 50+ файлов нарушают. Целевое состояние или пересмотр правила?
2. **GID-160** — разрешить `google.golang.org/grpc/codes` и `…/status` (маппинг ошибок ≠ транспорт)?
3. **GID-001** — исключение для замера длительности в метриках?
4. **Тесты** — применять ли GID-003 (uuid) и прочие правила к `_test.go`?
5. **GID-158** — что вносим в дефолтный allowlist: `schedule/`? `validate/`? пайплайн-папки event-сервисов?
6. **tools-cli / user-api** — исключать целиком через конфиг или чинить?
