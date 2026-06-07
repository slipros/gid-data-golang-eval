# Аудит DDD-конвенций backend-go: код → стайлгайд → линтер

Дата: 2026-06-07.
Метод: 6 параллельных обследований кода 19 сервисов `/mnt/w/GPM-Data/UDMP/backend-go`
по измерениям app/bootstrap · DAL · domain · transport · event+client · cross-cutting;
ключевые находки верифицированы по исходникам.
Направление обратное [FINDINGS.md](FINDINGS.md): там «правила → код» (нарушения и gap доков),
здесь «код → правила» — устойчивые конвенции (≥3 сервисов), которые могли не попасть
в skill `go-styleguide` и в линтер.

---

## 1. Сводка

Из ~40 устойчивых конвенций, наблюдаемых в коде:

- **большинство уже описано** в go-styleguide и/или покрыто правилами из [RULES.md](RULES.md)
  (операционные структуры — GID-210, build-сигнатуры — GID-212, sql.Null — GID-122,
  enum-конвертация — GID-143, ctx-helpers — GID-165/166/167, конструкторы event — GID-216,
  метрики — GID-174, конвертеры — GID-105/135 и т.д.);
- **~6 конвенций линтуемы, но отсутствуют** и в доках, и в линтере → §2 (кандидаты GID-218…GID-223);
- **~12 конвенций стоит описать в доках**, но линтовать нельзя/не нужно → §3;
- **аномалии**: новые нарушения существующих правил (§4.1) и межсервисные противоречия,
  требующие решения владельца стайлгайда до кодификации (§4.2).

---

## 2. Кандидаты в новые правила линтера (линтуемые пробелы)

| ID (предл.) | Слаг | Правило | Наблюдение | Линтуемость |
|---|---|---|---|---|
| GID-218 | search-pagination | Search-структуры `/domain/model`: пагинация — встроенное поле `Pagination` (value-тип `{Page, PageSize uint32}`), не указатели и не разбросанные поля | consent-api `internal/domain/model/model.go:4-10` + `attribute_group.go:54`, data-marketplace `pkg/datalab/domain/model/pagination.go`, organization-ticket; **нарушает lk-api** (`PageSize *uint32` россыпью, `organization/organizations.go:33-34`) | да: struct с суффиксом `Search` в model-слое |
| GID-219 | sort-struct | Сортировка — поле `SortBy` типа `<Entity>Sort {Field <Entity>SortField; Descending bool}`; enum `<Entity>SortField` с префиксом сущности и `String()`; разбросанные `SortField`/`SortDescending` поля запрещены | consent-api `attribute_group.go:34-54`, organization-ticket `organization_tickets.go:22-51`; **нарушает lk-api** (`type SortField string` без префикса, поля россыпью) | да |
| GID-220 | cache-wrapper-naming | Кэширующий репозиторий: имя `<Entity>WithCache`, оборачивает `*<Entity>` прямой ссылкой (поле-структура) — уточнение GID-159 | consent-api `dal/repository/webhook_with_cache.go:13`, `organization_configuration_with_cache.go:13` | да: тип `*WithCache` в `/dal/repository` |
| GID-221 | count-method-naming | Метод подсчёта — суффикс `Count` (`<Entities>Count`, `CommentsCount`); префикс `Total*` запрещён — расширение GID-114 `gidentitymethod` | ~10 методов `*Count` (organization-ticket, organization-portfolio, consent-api) против 3 `Total*` только в ClickHouse-репо consent-api (`clickhouse_consent.go:21`) | да |
| GID-222 | event-schema-func | Schema registry в event-слое — standalone-функция `New<Event>Schema()`, не метод consumer'а/producer'а — расширение GID-216 `gideventctor` | все сервисы; **нарушает event-packer** (`internal/event/consumer/kafka/http_request.go:59` — метод `Schema()`) | да |
| GID-223 | op-struct-upsert | Расширение GID-210: распознавать `Upsert<X>`/`Delete<X>` как операционные структуры (Delete — минимальный набор ключей; Upsert — поля INSERT/UPDATE) | consent-api `dal/entity/webhook.go:81-106` (UpsertWebhook, DeleteWebhook), organization-ticket | да, после фиксации допустимых полей |

---

## 3. Пробелы доков (описать в go-styleguide; линтер — нет или потом)

| Конвенция | Наблюдение | Куда |
|---|---|---|
| Browse/Lightweight-типы: `<Entity>Browse` — строка пагинированной выборки, `<Entity>Lightweight` — облегчённое представление | consent-api `dal/entity/webhook.go:75` и model-слой, organization-ticket `organization_ticket_comment.go:10,33`, organization-portfolio | entity.md, model.md (+ решение по терминологии, §4.2) |
| List-результаты: `<Entities>List {Data, PageSize, PageNumber, TotalCount}` либо `ListMeta {Pagination, TotalCount}` | lk-api `organization/organizations.go:38-44`, data-marketplace `pagination.go:13-15` | model.md (+ решение, §4.2) |
| Слайс-типы с методами-экстракторами: `IDs()`, `Values()`, `Filter(FilterFunc)` | consent-api `model/profile.go:53-88`, lk-api `organizations.go:68-77` | model.md |
| Маппинг domain-ошибок → коды транспорта: `switch { case errors.Is(...) }` в теле хендлера; без централизованного маппера | consent-api grpc-хендлеры, lk-api `pkg/account/server/http/handler/login.go:36-49`, event-api | server.md |
| Глобальный `WithErrorConverters` для `validator.Result` → `codes.InvalidArgument` через `grpcerror.ValidationErrorFromV2Validator` | 7 gRPC-сервисов, consent-api `app/api/application.go:445-454` | server.md / app.md |
| Роутер-обёртка `NewHandlerFunc[T]` / `NewHandlerFuncWithValidate[T]`: parse (roamer) → validate → 422 → handler | consent-api `server/http/router/handler.go:28-93`, lk-api, event-api; file-storage делает иначе (`ExtendedHandlerFunc`) | server.md |
| HTTP-сервер: system router (`NewSystemRouterWithConnectionsPings`) + application routers | consent-api `application.go:264-293`, event-api, organization-ticket | app.md |
| Метрики создаются и регистрируются первым шагом `New()` | все API-сервисы | app.md |
| Версионирование событий v1/v2: `oneof payload` + отдельные handler/валидаторы на версию, либо отдельные producer-типы | gid-sso-consent-consumer `consent_event.go:71-95`, event-api | event.md |
| Узкий interface обработчика в consumer (`<Event>Handler` / `<Entity>Service` только с нужными методами) | 4/4 обследованных consumer'ов | event.md (формализация; частично гарантируется GID-197) |
| Обработка ошибок consumer'а по категориям: convert/validate → `Commit` (skip+log), retriable → `Rollback`, exhausted → `Commit` | consent-webhook-sender `webhook_trigger.go:71-111`, gid-sso-consent-consumer, event-packer | event.md |
| Client: defer-замер метрик в каждом внешнем методе; «ожидаемое отсутствие» — значение, не ошибка (`Exists() (bool, error)` → `(false, nil)`) | file-storage `client/minio.go:100-120` | client.md |
| Интеграционные тесты: полный harness `/test/integration` (testcontainers, `//go:build integration`) vs `*_integration_test.go` рядом с кодом | consent-api (harness) vs data-reports, data-marketplace, lk-api (inline) | новый раздел (+ решение, §4.2) |
| Шедулинг: `internal/schedule/schedule.go` — wrapper над `gdscheduler`, Job'ы — `job/`-пакеты модулей | consent-api, data-marketplace, user-api; file-storage кладёт в `internal/job/` | app.md + связка с allowlist GID-158 (FINDINGS.md §1.5) |

Не линтуемые и достаточно описанные (bootstrap-последовательность main.go, структура `New()`,
graceful shutdown `gdapp.Application{Servers, Closers}`, CLI-флаги Destination/Action,
env-префиксы) — подтверждены кодом, расхождений с app.md нет, кроме аномалий §4.1.

---

## 4. Аномалии

### 4.1 Новые нарушения существующих правил (сверх FINDINGS.md §1)

| Сервис | Нарушение | Правило | Файл |
|---|---|---|---|
| data-reports | ctx-helpers (`ContextKey`, `ContextWithMemberID`, `MemberIDFromContext`) живут в `/dal/entity` вместо `/domain/model` | GID-165/166/167 | `internal/dal/entity/context.go:9-22` |
| lk-api, user-api | пакет `internal/metrics` вместо `internal/metric` (остальные 17 — `metric`) | GID-174 | `lk-api/internal/metrics/`, `user-api/internal/metrics/` |
| file-storage | composition root называется `NewApplication`, не `New` | GID-104/конвенция app | `internal/app/file-storage/application.go:41` |
| file-storage | enum inline в entity-файле вместо `/dal/entity/enum` | GID-211 | `internal/dal/entity/object.go:9-22` |
| organization-ticket | enum inline в entity-файле | GID-211 | `internal/dal/entity/organization_ticket.go:10-43` |
| file-storage | указатели (`*string`, `*uint64`, `*time.Time`) в entity вместо `sql.Null*` | GID-122 | `internal/dal/entity/upload.go:23-24`, `object.go:32-47` |
| lk-api | логгер в `domain/service` (поле `logger *logrus.Entry` у сервиса) — единственный сервис | конвенция (не кодифицирована) | `internal/domain/service/member.go:18-28` |
| lk-api | в `main.go` logger инициализируется до Sentry/Trace (нарушение bootstrap-порядка app.md) | не линтуется | `cmd/lk-api/main.go:64-115` |

Списки — кандидаты в `settings.exclude` либо техдолг к починке (как FINDINGS.md §1).

### 4.2 Межсервисные противоречия — нужно решение владельца стайлгайда

1. **Пагинация в Search**: значения `uint32` (большинство) vs указатели `*uint32` (lk-api) → блокирует GID-218.
2. **Сортировка**: отдельная структура `<Entity>Sort` (большинство) vs поля россыпью (lk-api) → блокирует GID-219.
3. **`Count` vs `Total*`**: суффикс `Count` доминирует; `Total*` — только ClickHouse-репо consent-api → блокирует GID-221.
4. **Browse vs Lightweight**: оба термина живут рядом без зафиксированного различия назначений.
5. **List-результат**: пагинация полями в `<Entities>List` (lk-api) vs отдельный `ListMeta` (data-marketplace).
6. **Кэш-библиотеки**: `lru` и `gdhelper.Cache` параллельно в одном сервисе (consent-api `*_with_cache.go`) — выбрать стандарт.
7. **Имена методов Metrics-интерфейса client-слоя**: `ObserveRequest` (доки) vs `AddStorageProviderRequest` (file-storage) vs `IncrementClientRequest` (gid-sso-consent-consumer) — нет стандарта.
8. **Логгер в domain/service**: запретить (как в 18 сервисах) или разрешить (lk-api)?
9. **Интеграционные тесты**: harness `/test/integration` vs `*_integration_test.go` — что норма?
10. **`Produce`**: слайс (доки) vs variadic (event-api `http_request.go:29`) — допустить оба?

### 4.3 Не подтвердилось при верификации

- «event-collector не регистрирует группу метрик StorageProvider» — ложь: `Register` вызывается
  явно для всех групп (`internal/metric/prometheus.go:47-69`), поля StorageProvider там нет.
- «consent-webhook-sender нарушает GID-105 именем `WebhookTriggerFromEvent`» — спорно: имя
  матчится паттерном `<Dst>From<Src>`; проверить фактический вердикт `gidconvnaming` перед записью в техдолг.

---

## 5. Порядок работ

1. Решить вопросы §4.2 п.1-3 → реализовать GID-218, GID-219, GID-221 (процесс из RULES.md).
2. GID-220, GID-222, GID-223 — независимы от решений, можно брать сразу.
3. Дополнить доки go-styleguide по §3 (entity.md, model.md, server.md, app.md, event.md, client.md).
4. Нарушения §4.1 — внести в backlog исключений/техдолга вместе с FINDINGS.md §1.
