# Audit of backend-go DDD conventions: code → styleguide → linter

Date: 2026-06-07.
Method: 6 parallel code surveys of 19 services in `/mnt/w/GPM-Data/UDMP/backend-go`
along the dimensions app/bootstrap · DAL · domain · transport · event+client · cross-cutting;
key findings verified against the sources.
The direction is the reverse of [FINDINGS.md](FINDINGS.md): there it is "rules → code" (violations and doc gaps),
here it is "code → rules" — stable conventions (≥3 services) that may not have made it
into the skill `go-styleguide` and the linter.

---

## 1. Summary

Out of ~40 stable conventions observed in the code:

- **most are already described** in go-styleguide and/or covered by rules from [RULES.md](RULES.md)
  (operational structs — GID-210, build signatures — GID-212, sql.Null — GID-122,
  enum conversion — GID-143, ctx helpers — GID-165/166/167, event constructors — GID-216,
  metrics — GID-174, converters — GID-105/135, etc.);
- **~6 conventions are lintable but absent** from both the docs and the linter → §2 (candidates GID-218…GID-223);
- **~12 conventions are worth describing in the docs**, but cannot/should not be linted → §3;
- **anomalies**: new violations of existing rules (§4.1) and cross-service contradictions
  that require a decision from the styleguide owner before codification (§4.2).

---

## 2. Candidates for new linter rules (lintable gaps)

| ID (proposed) | Slug | Rule | Observation | Lintability |
|---|---|---|---|---|
| GID-218 | search-pagination | Search structs in `/domain/model`: pagination is an embedded `Pagination` field (value type `{Page, PageSize uint32}`), not pointers and not scattered fields | consent-api `internal/domain/model/model.go:4-10` + `attribute_group.go:54`, data-marketplace `pkg/datalab/domain/model/pagination.go`, organization-ticket; **violated by lk-api** (scattered `PageSize *uint32`, `organization/organizations.go:33-34`) | yes: struct with the `Search` suffix in the model layer |
| GID-219 | sort-struct | Sorting — a `SortBy` field of type `<Entity>Sort {Field <Entity>SortField; Descending bool}`; the `<Entity>SortField` enum with an entity prefix and `String()`; scattered `SortField`/`SortDescending` fields are forbidden | consent-api `attribute_group.go:34-54`, organization-ticket `organization_tickets.go:22-51`; **violated by lk-api** (`type SortField string` without a prefix, scattered fields) | yes |
| GID-220 | cache-wrapper-naming | Caching repository: named `<Entity>WithCache`, wraps `*<Entity>` by direct reference (struct field) — a refinement of GID-159 | consent-api `dal/repository/webhook_with_cache.go:13`, `organization_configuration_with_cache.go:13` | yes: a `*WithCache` type in `/dal/repository` |
| GID-221 | count-method-naming | Count method — the `Count` suffix (`<Entities>Count`, `CommentsCount`); the `Total*` prefix is forbidden — an extension of GID-114 `gidentitymethod` | ~10 `*Count` methods (organization-ticket, organization-portfolio, consent-api) versus 3 `Total*` only in the consent-api ClickHouse repo (`clickhouse_consent.go:21`) | yes |
| GID-222 | event-schema-func | Schema registry in the event layer — a standalone function `New<Event>Schema()`, not a method of the consumer/producer — an extension of GID-216 `gideventctor` | all services; **violated by event-packer** (`internal/event/consumer/kafka/http_request.go:59` — a `Schema()` method) | yes |
| GID-223 | op-struct-upsert | Extension of GID-210: recognize `Upsert<X>`/`Delete<X>` as operational structs (Delete — a minimal set of keys; Upsert — INSERT/UPDATE fields) | consent-api `dal/entity/webhook.go:81-106` (UpsertWebhook, DeleteWebhook), organization-ticket | yes, once the allowed fields are fixed |

---

## 3. Doc gaps (describe in go-styleguide; linter — no or later)

| Convention | Observation | Where |
|---|---|---|
| Browse/Lightweight types: `<Entity>Browse` — a row of a paginated selection, `<Entity>Lightweight` — a lightweight representation | consent-api `dal/entity/webhook.go:75` and the model layer, organization-ticket `organization_ticket_comment.go:10,33`, organization-portfolio | entity.md, model.md (+ terminology decision, §4.2) |
| List results: `<Entities>List {Data, PageSize, PageNumber, TotalCount}` or `ListMeta {Pagination, TotalCount}` | lk-api `organization/organizations.go:38-44`, data-marketplace `pagination.go:13-15` | model.md (+ decision, §4.2) |
| Slice types with extractor methods: `IDs()`, `Values()`, `Filter(FilterFunc)` | consent-api `model/profile.go:53-88`, lk-api `organizations.go:68-77` | model.md |
| Mapping domain errors → transport codes: `switch { case errors.Is(...) }` in the handler body; no centralized mapper | consent-api grpc handlers, lk-api `pkg/account/server/http/handler/login.go:36-49`, event-api | server.md |
| Global `WithErrorConverters` for `validator.Result` → `codes.InvalidArgument` via `grpcerror.ValidationErrorFromV2Validator` | 7 gRPC services, consent-api `app/api/application.go:445-454` | server.md / app.md |
| Router wrapper `NewHandlerFunc[T]` / `NewHandlerFuncWithValidate[T]`: parse (roamer) → validate → 422 → handler | consent-api `server/http/router/handler.go:28-93`, lk-api, event-api; file-storage does it differently (`ExtendedHandlerFunc`) | server.md |
| HTTP server: system router (`NewSystemRouterWithConnectionsPings`) + application routers | consent-api `application.go:264-293`, event-api, organization-ticket | app.md |
| Metrics are created and registered as the first step of `New()` | all API services | app.md |
| Event versioning v1/v2: `oneof payload` + separate handlers/validators per version, or separate producer types | gid-sso-consent-consumer `consent_event.go:71-95`, event-api | event.md |
| A narrow handler interface in the consumer (`<Event>Handler` / `<Entity>Service` with only the needed methods) | 4/4 surveyed consumers | event.md (formalization; partially guaranteed by GID-197) |
| Consumer error handling by category: convert/validate → `Commit` (skip+log), retriable → `Rollback`, exhausted → `Commit` | consent-webhook-sender `webhook_trigger.go:71-111`, gid-sso-consent-consumer, event-packer | event.md |
| Client: deferred metric measurement in every external method; "expected absence" is a value, not an error (`Exists() (bool, error)` → `(false, nil)`) | file-storage `client/minio.go:100-120` | client.md |
| Integration tests: a full harness `/test/integration` (testcontainers, `//go:build integration`) vs `*_integration_test.go` next to the code | consent-api (harness) vs data-reports, data-marketplace, lk-api (inline) | new section (+ decision, §4.2) |
| Scheduling: `internal/schedule/schedule.go` — a wrapper over `gdscheduler`, Jobs — `job/` packages of modules | consent-api, data-marketplace, user-api; file-storage puts it in `internal/job/` | app.md + link to the GID-158 allowlist (FINDINGS.md §1.5) |

Non-lintable and sufficiently described items (main.go bootstrap sequence, `New()` structure,
graceful shutdown `gdapp.Application{Servers, Closers}`, CLI flags Destination/Action,
env prefixes) — confirmed by the code, no divergences from app.md except for the anomalies in §4.1.

---

## 4. Anomalies

### 4.1 New violations of existing rules (beyond FINDINGS.md §1)

| Service | Violation | Rule | File |
|---|---|---|---|
| data-reports | ctx helpers (`ContextKey`, `ContextWithMemberID`, `MemberIDFromContext`) live in `/dal/entity` instead of `/domain/model` | GID-165/166/167 | `internal/dal/entity/context.go:9-22` |
| lk-api, user-api | package `internal/metrics` instead of `internal/metric` (the other 17 use `metric`) | GID-174 | `lk-api/internal/metrics/`, `user-api/internal/metrics/` |
| file-storage | composition root is named `NewApplication`, not `New` | GID-104/app convention | `internal/app/file-storage/application.go:41` |
| file-storage | enum inline in the entity file instead of `/dal/entity/enum` | GID-211 | `internal/dal/entity/object.go:9-22` |
| organization-ticket | enum inline in the entity file | GID-211 | `internal/dal/entity/organization_ticket.go:10-43` |
| file-storage | pointers (`*string`, `*uint64`, `*time.Time`) in entity instead of `sql.Null*` | GID-122 | `internal/dal/entity/upload.go:23-24`, `object.go:32-47` |
| lk-api | a logger in `domain/service` (a `logger *logrus.Entry` field on the service) — the only such service | convention (not codified) | `internal/domain/service/member.go:18-28` |
| lk-api | in `main.go` the logger is initialized before Sentry/Trace (violates the app.md bootstrap order) | not lintable | `cmd/lk-api/main.go:64-115` |

These lists are candidates for `settings.exclude` or tech debt to fix (as in FINDINGS.md §1).

### 4.2 Cross-service contradictions — a styleguide owner decision is needed

1. **Pagination in Search**: `uint32` values (the majority) vs `*uint32` pointers (lk-api) → blocks GID-218.
2. **Sorting**: a dedicated `<Entity>Sort` struct (the majority) vs scattered fields (lk-api) → blocks GID-219.
3. **`Count` vs `Total*`**: the `Count` suffix dominates; `Total*` — only the consent-api ClickHouse repo → blocks GID-221.
4. **Browse vs Lightweight**: both terms coexist with no recorded distinction in purpose.
5. **List result**: pagination as fields in `<Entities>List` (lk-api) vs a separate `ListMeta` (data-marketplace).
6. **Cache libraries**: `lru` and `gdhelper.Cache` in parallel within one service (consent-api `*_with_cache.go`) — pick a standard.
7. **Method names of the client-layer Metrics interface**: `ObserveRequest` (docs) vs `AddStorageProviderRequest` (file-storage) vs `IncrementClientRequest` (gid-sso-consent-consumer) — no standard.
8. **Logger in domain/service**: forbid (as in 18 services) or allow (lk-api)?
9. **Integration tests**: the `/test/integration` harness vs `*_integration_test.go` — which is the norm?
10. **`Produce`**: a slice (docs) vs variadic (event-api `http_request.go:29`) — allow both?

### 4.3 Not confirmed during verification

- "event-collector does not register the StorageProvider metric group" — false: `Register` is called
  explicitly for all groups (`internal/metric/prometheus.go:47-69`), and there is no StorageProvider field there.
- "consent-webhook-sender violates GID-105 with the name `WebhookTriggerFromEvent`" — debatable: the name
  matches the `<Dst>From<Src>` pattern; check the actual `gidconvnaming` verdict before recording it as tech debt.

---

## 5. Work order

1. Resolve §4.2 items 1-3 → implement GID-218, GID-219, GID-221 (the process from RULES.md).
2. GID-220, GID-222, GID-223 — independent of the decisions, can be taken right away.
3. Extend the go-styleguide docs per §3 (entity.md, model.md, server.md, app.md, event.md, client.md).
4. The violations in §4.1 — add to the exclusions/tech-debt backlog together with FINDINGS.md §1.
