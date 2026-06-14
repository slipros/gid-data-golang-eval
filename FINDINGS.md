# Audit: exclusions in backend-go and missed rules

Date: 2026-06-07.
Sources: [RULES.md](RULES.md), skill `go-styleguide` (full gap analysis of the docs),
scan of 20 services in `/mnt/w/GPM-Data/UDMP/backend-go`.
The reverse direction (conventions from code that did not make it into the styleguide) — [FINDINGS_DDD.md](FINDINGS_DDD.md).

This document records: (1) places where real code violates already implemented rules —
candidates for `settings.exclude` or a signal to relax the rule; (2) rules missing
from the registry, with a decision for each; (3) open questions.

---

## 1. Exclusions — violations of implemented (✅) rules

### 1.1 GID-001 `no-time-now` — ~15 occurrences, three clusters

| Cluster | Files |
|---|---|
| **Metric measurement** (`defer calculateMetrics(..., time.Now())`) | `file-storage/internal/client/minio.go:101,123,172,225,251` · `profile-targeting-api/internal/client/http/metrics.go:8` · `file-storage/internal/app/file-storage/application.go:212` |
| **`nowFunc` in validators** (max date = "now") | `consent-api/pkg/consent/v{1,2}/server/http/handler/validate/statuses.go:35` · `event-api/internal/server/http/router/handler/validate/consent_event_v{1,2}.go:90` · `gid-sso-consent-consumer/internal/event/consumer/kafka/validate/consent_event_v1.go:80,215`, `consent_event_v2.go:96,484` |
| **Schedulers** | `consent-api/internal/schedule/schedule.go:20` · `data-marketplace/internal/schedule/schedule.go:24` · `data-marketplace/pkg/marketplace/job/materialization/job.go:144` |

**Question**: relax the rule for duration measurement in metric code — or put everything into `settings.exclude`.

### 1.2 GID-112 `create-update-no-return` — ~12 occurrences

| Service | File | Method |
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

A ready-made list for `settings.exclude` / `//nolint:gidcreateupdate` — or tech debt to fix.

### 1.3 GID-144/145 `errors-in-model/entity` — the most widespread (50+ files)

`errors.go` with `errors.New` lives in `/domain/service/` and `/dal/repository/` almost everywhere:

- **lk-api** — 30+ files: `internal/domain/service/errors.go`, `internal/validate/errors.go`, `pkg/{account,member,consent,organization,…}/domain/service/errors.go`, `pkg/*/dal/repository/errors.go`
- **organization-portfolio** — `internal/domain/service/errors.go`, `internal/dal/repository/errors.go`
- **organization-ticket** — `internal/domain/service/errors.go:6-12`
- **file-storage** — `internal/service/errors.go` (26 errors; the folder itself violates GID-158), `internal/client/errors.go`, `internal/dal/repository/errors.go`
- **event-packer** — `internal/domain/service/errors.go:9`
- **gid-sso-consent-consumer** — `internal/event/consumer/kafka/validate/errors.go`

**Key question**: is the rule the target state (this is tech debt), or should the team norm be reconsidered.

### 1.4 GID-160 `grpc-via-repository` — a false-positive class

In `/domain/service`, only `google.golang.org/grpc/codes` and `…/status` are imported
for **error code mapping**, not for transport:
`gid-sso-consent-consumer/internal/domain/service/user_v2.go:10-11`,
`lk-api/internal/domain/service/{file,file_upload,organizations}.go`.

**Proposal**: allow `codes`/`status` in the rule — otherwise a flood of nolint.

### 1.5 GID-158 `dir-tree` — recurring "non-standard" folders

| Folder | Services | Assessment |
|---|---|---|
| `schedule/` | consent-api, data-marketplace | candidate for the default allowlist |
| `validate/` | file-storage, lk-api, consent-api | to discuss (general-purpose validation) |
| `statement/`, `batch/`, `collector/`, `consumer/` | event-collector | specialized pipeline |
| `enricher/` | event-enricher | specialized pipeline |
| `job/`, `producer/`, `service/` | file-storage | `service/` — a violation (should be `domain/service`) |
| `metrics/` (instead of `metric`), `testharness/` | lk-api | `metrics` → rename |

### 1.6 Candidates for exclusion as whole modules

- **tools-cli** — a CLI utility (urfave/cli) without layers: exclude from GID-110/111/130s/132/151/158.
- **user-api** — legacy structure (`/internal/service/{account,member,…}` instead of `/domain/service`); violates GID-132, GID-148 (`member` → `organizationService`, `auditLog`, `providerService`; `account` → `memberService` + repository accessed directly), GID-158.
- **test/**, **example/** (event-collector: `example/http-request-producer/producer.go:8` imports `github.com/google/uuid` → GID-137).

### 1.7 Miscellaneous

- **GID-123, real-world case**: event-collector `internal/domain/model/enum/{consent_event_type,consent_event_v2_type,device_type}.go` — `type X = string` (alias instead of a named type). Taking it as the positive case for the GID-123 eval.
- **GID-003 in tests**: `uuid.Must(uuid.NewV4())` is widespread in `_test.go` (consent-webhook-trigger — 30+, data-reports). Question: should the rule apply to tests.

---

## 2. Missed rules

### 2.1 Accepted for work (to be implemented in this repo)

| ID | Slug | Rule | Source |
|---|---|---|---|
| GID-123 | enum-string-based | Implementation of the existing 🔜: named string type; ban on alias (`type X = string`) and int enums; ban on groups of untyped string consts in model/entity | styleguide.md#enum + the event-collector case |
| GID-168 | no-db-tags-in-model | `db:` tags on struct fields are forbidden in `/domain/**` — DB mapping lives in entity | model.md |
| GID-169 | errors-in-error-file | Refinement of GID-144/145: error variables in `/domain/model` and `/dal/entity` live in the file `error.go`/`errors.go` | model.md, entity.md |
| GID-170 | no-event-import | `/domain/**` and `/dal/**` do not import `/event/**` — the event layer converts model ↔ DTO, not the other way around | event.md, ARCHITECTURE.md |
| GID-171 | filter-location | Filter structs: in `/dal/**` — only `/dal/entity/filter`; in `/domain/**` — only the model layer | model.md, entity.md |
| GID-172 | client-no-entity | `/client/**` does not import `/dal/**` — the client has its own types | client.md |
| GID-173 | iface-entity-prefix | Dependency interfaces carry an entity prefix (`HelloRepository`); bare roles (`Repository`, `Service`, …) are forbidden | styleguide.md#interfaces |
| GID-174 | metric-prometheus-struct | Package `/metric`: type `Prometheus` (aggregator of metrics by protocol) with a `Register` method; the package is named `metric` | convention across all services |
| GID-175 | in-transaction | Transaction types (`InTransactionFunc`, `InTransactionWithReturnFunc[T]`) live in `/domain/model`; service/usecase use the named type (a connection with the tx signature is injected via the constructor); repo/service do not declare tx methods | requirement of 2026-06-07 |

Canonical form of GID-175 (lives in `/domain/model`):

```go
type InTransactionFunc func(ctx context.Context, fn func(ctx context.Context) error) error

type InTransactionWithReturnFunc[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)

// NewInTransactionWithReturnFunc wraps InTransactionFunc to support a return value.
func NewInTransactionWithReturnFunc[T any](tx InTransactionFunc) InTransactionWithReturnFunc[T] { … }
```

### 2.2 Already covered by existing rules (no implementation needed)

- **transport-converter-naming** (`Model<T>FromGRPC` / `GRPC<T>FromModel`) — `gidconvnaming` (GID-105/135) already applies in the `server` and `event` layers; the `<Dst><Type>From<Src>` pattern covers transport names. The only thing not covered is the naming of *private* enum helpers (`grpc<X>FromEnum`) — low value, deferred.
- **dal-enum-string** — GID-124 (`gidenumstring`) is not limited to a layer and already requires `String()` on any string enums, including `/dal/entity/enum`; "named type, not alias" is closed by the implementation of GID-123.

### 2.3 Contradiction found — nested Options

From the docs (app.md) one can derive "Options do not contain nested `*Options`", but real code
does the opposite everywhere: `Options` in `internal/app/api/options.go` aggregates
`GRPCOptions`, `KafkaOptions`, `TraceOptions` (consent-api, event-api, event-collector,
organization-ticket, …). Decision: candidate rejected; when reworking GID-126, take into
account that Options composition in the app layer is the norm.

### 2.4 Deferred (medium value, not taken now)

| Slug | Rule | Why deferred |
|---|---|---|
| ctor-deps-signature | The event-layer part (consumer accepts logger, producer does not) **is implemented as GID-216** (2026-06-07). The remainder — Client must accept Metrics (client.md) — is deferred until the client-layer spec is clarified | narrow, deterministic, but few violations |
| slice-type-required | A slice type is mandatory for the main entity (`Jobs []Job`) | no reliable indicator of the "main" entity — FP |
| logger-nil-default | Constructor with logger: nil → `logrus.StandardLogger()` | not recorded in the docs, only in the code of 3 services |
| mocks-location | Mocks in `mock*/` subpackages | a convention from code, not in the styleguide |
| private-enum-converter-naming | `grpc<X>FromEnum`, `avro<X>FromDTO` for private converters | low value |

### 2.5 Remains for code review (not statically checkable)

Validator patterns (`When` vs callback, `NewNested`+`NewEach`, internal vs public validation,
proto3 enums via `NewInRange`, `NewTime` only for string) · "a converter is a pure function" ·
"a handler contains no business logic" · "a usecase depends on services, not on repos" (partially
closed by GID-132) · FSM transition map + `CanTransitionTo` · graceful shutdown in main ·
ID/CreatedAt generation in the converter · operational entity structs with a minimal field set.

---

## 3. Open questions (require a decision from the styleguide owner)

1. **GID-144/145 vs reality** — 50+ files violate. Target state or rule revision?
2. **GID-160** — allow `google.golang.org/grpc/codes` and `…/status` (error mapping ≠ transport)?
3. **GID-001** — an exception for duration measurement in metrics?
4. **Tests** — should GID-003 (uuid) and other rules apply to `_test.go`?
5. **GID-158** — what goes into the default allowlist: `schedule/`? `validate/`? the pipeline folders of event services?
6. **tools-cli / user-api** — exclude entirely via config or fix?
