# Styleguide ↔ linter sync audit (2026-07-04)

Full reconciliation of the `go-styleguide` skill docs (source of truth) against
RULES.md and `.golangci.yml`. Method: 4 parallel reviewers, one per doc group
(core styleguide / model+dal+domain / transport+convert / architecture+app+libs),
each doing a two-way check: doc requirement → linter coverage, and rule `Source`
→ doc confirmation.

**Verdict:** coverage is solid (~110 normative requirements traced to active
rules, all claimed linters verified as enabled in `.golangci.yml`), no rule
tightens beyond the docs. Found: **2 hard contradictions**, a set of
doc-sync defects (rule deviates from doc without the doc saying so), and
several automatable gaps — including the missing **convert-purity** rule.

Every discrepancy below needs a decision: fix the doc (if the linter reflects a
newer decision) or fix the rule (if the doc is right). Per the project rule,
divergence is a tooling defect — fix at the source.

---

## 1. Hard contradictions (linter would flag the doc's own canon)

| # | Where | Doc says | Linter says |
|---|---|---|---|
| C1 | GID-228 vs ARCHITECTURE.md "Поток зависимостей" | "Клиенты используются репозиториями…", "допускается прямая зависимость domain → client (минуя dal)" | `/domain/**` and `/dal/**` must not import `/client/**` (decision 2026-06-07) | 
| C2 | GID-167 vs model.md (ctx keys) | canonical example `type ContextKey uint8` + `iota` | requires `type ContextKey string` + snake_case const values |

Both: the rule's source is a later decision; the doc was never updated. Decide
which side is canon, then fix the loser.

## 2. Rule vs doc-example conflicts

- **GID-176 / event producer**: event.md §producer says errors are wrapped
  "аналогично repository и client" and example_event.md uses `errors.Wrap` —
  but GID-176 treats only `/client/**` and `/dal/repository` as the boundary and
  *bans* `Wrap` on incoming non-static errors inside `/domain/**`. `/event` is
  in neither list. Either extend the boundary to producer code or rewrite the
  example to `WithMessage`.
- **GID-124 String() on every string enum**: model.md/example_model.md show a
  model enum with only `CanTransitionTo`, no `String()` — a strictly
  by-the-book model enum gets flagged. Either add `String()` to the doc canon
  or scope GID-124 to `/dal/entity/enum`.
- **model.md "model — один пакет без подпакетов"** vs GID-132/171 explicitly
  treating `/domain/model/*` (filter, enum) as a full model layer. Doc is
  stricter than the linter; sync the doc (subpackages are in active use).

## 3. Doc-sync defects (rule weakens/deviates, doc silent)

- **GID-133**: doc (styleguide.md §Структура пакетов) has no shared-helper
  exception and no 3-layer scoping — both were added in the linter only.
- **GID-103**: `v`/`h` receiver exceptions are justified by validator.md /
  server.md examples but absent from styleguide.md §Именование ресиверов.
- **GID-126**: the "bare `Options` allowed in app layer" exception (FINDINGS
  §2.3) is not reflected in styleguide.md §Options-паттерн.
- **GID-169**: claims "`err.go` is the canonical name from entity.md" — entity.md
  and example_entity.md actually write `error.go`; `err.go` appears nowhere.
- **GID-113**: content is right but the `Source` anchor is wrong — the rule text
  lives in styleguide.md §Сигнатуры методов, not #options-pattern.
- **ARCHITECTURE.md names `err113`** as the enforcing linter for runtime
  `errors.New`, but the config mutes err113's dynamic-errors check; the actual
  enforcement is GID-136/146. Fix the doc reference.
- **`.golangci.yml` description of `gidoptsstyle`** still says "opts …
  *embedded* in the struct" — stale after the GID-152 inversion (embedding is
  now forbidden, named `opts` field required).

## 4. Coverage narrower than the doc (same rule, smaller net)

- **GID-233**: converter.md bans *any* direct enum cast including
  `string(in.Type)`; the rule only catches enum↔enum across packages.
- **GID-215**: service.md bans any inline conversion; the rule only catches
  non-empty composite literals of entity types (inline field-by-field mapping
  without a literal passes).
- **GID-121**: model.md's principle is "zero value expresses absence" in
  general; the rule checks only `*time.Time` and string-kind pointers
  (`*int`/`*float` pass). Deliberate FP trade-off — document it in RULES.md.

## 5. Automatable gaps (new-rule candidates)

1. **Convert purity (top priority, requested 2026-07-04).** converter.md:
   "Конвертеры — функции без побочных эффектов". No rule controls what a
   `convert/` package may import/call; the layer matrix only shields transport
   convert (GID-224), and event convert has no shield at all. Canonical
   examples import exactly: stdlib (`time`), `gofrs/uuid`, `gdhelper`,
   `gderror`, `pkg/errors`, plus vocabulary packages (model, entity, dto,
   enum, genproto/pb, queue+message for producer convert) — never
   service/usecase/repository/client/logrus.
   **Sketch (`gidconvpure`)**: in packages whose path segment is `convert`,
   allow imports of stdlib + vocabulary layers (`/domain/model/**`,
   `/dal/entity/**`, `/event/dto`, generated proto) + a configurable utility
   allowlist (defaults: pkg/errors, gofrs/uuid, gdhelper, gderror, queue
   message types); ban everything else (service, usecase, repository, client,
   metric, logrus, net/http, database/sql, …). Deterministic, import-level.
2. **`errors.WithMessage` ban in `/domain/service`** (service.md L62: WithMessage
   is usecase-only). Complement to GID-176.
3. **Client method signatures** — extend GID-111 scope to `/client/**`
   (client.md §Сигнатуры методов; input by pointer, output by value).
4. **`event/dto` follows model rules** — extend GID-121/123 scope to
   `/event/dto` (event.md L48).
5. **Constructor returns bare `*T`, no `error`** (repository.md L11) — check
   `New*` in service/usecase/repository: single pointer result.
6. **module.md `common` import-alias prefix** for `internal/domain` imports from
   `pkg/{module}`; also GID-224's `/internal/` boundary assumption breaks for
   the `pkg/{module}` layout — transport bans silently off there.
7. Minor: urfave/cli flag naming (kebab-case Name, UPPER_SNAKE env), producer
   error wrapping (folds into #1/GID-176 fix), table name as const.

## 6. repo→service→usecase hierarchy — enforcement status

Doc statements: service works with exactly one entity (service.md L3); usecase
never depends on repositories directly, data access goes through services
(usecase.md L24); service calls its own entity's repository "если явно не
указано иное" (service.md L15 — a soft default).

- **usecase → repository directly: enforced.** GID-132 bans `/domain/usecase`
  → `/dal/**` imports, and a repo interface needs entity types, whose import is
  also banned. (Residual loophole: an interface returning model types is
  indistinguishable from a service — but that *is* a service, layer-wise.)
- **repository called only from service: effectively enforced.** Transport
  can't (GID-224), usecase can't (GID-132), metric can't (GID-226).
- **service = one entity: NOT enforced.** GID-148 only bans service→service
  fields; one service injecting repositories of several entities passes.
  Deterministic "ownership" of a repository can't be established without
  FP-heavy heuristics; the doc itself hedges (L15) — leave on review, note in
  RULES.md "not portable".
- **usecase orchestrates many services: by design**, GID-148 allows it.

---

*Reviewers' raw reports: 4 agent runs, 2026-07-04 session. Registry lines
referenced against RULES.md and `.golangci.yml` at commit 4b6dfb9.*

---

## Decision log (2026-07-04, owner)

- **C1 (GID-228)**: both patterns are canon — a client is used by a repository
  (client models → entity) *or* directly by a service (model ↔ client models;
  the service API always takes/returns model). Docs (ARCHITECTURE.md,
  service.md) rewritten; GID-228 narrowed to `/domain/usecase` → `/client` only.
- **C2 (GID-167)**: all enums are always string-based → the linter is right;
  model.md `ContextKey uint8`+`iota` example replaced with `string` + snake_case.
- **Gap #1 (convert purity)**: implemented as **GID-235 `gidconvpure`**.
- **Hierarchy "service = one entity"**: implemented as **GID-236
  `gidserviceentity`** (deterministic core: same-package `*Repository`
  interface fields must match the service's entity; escape hatch per
  service.md "если явно не указано иное").
- Stale `gidoptsstyle` description in `.golangci.yml` fixed.
- Remaining open items (§2-§5) — pending separate decisions.

## Decision log, round 2 (2026-07-04, owner) — ALL remaining items resolved

- §2 GID-176: reworked (v2) — errors from **external calls** (another module,
  incl. stdlib) are wrapped with `errors.Wrap` in ANY layer; interface-call
  boundary extended to `/event/**`; the `/domain` Wrap-ban now applies to
  same-module errors only. New **GID-237 `gidwithmessage`** bans
  `errors.WithMessage` in `/domain/service` (usecase-only).
- §2 GID-124: `String()` stays mandatory on every enum — model.md canon updated.
- §2 model subpackages: allowed (top-down deps) — model.md updated.
- §3 GID-133/103/126/169/113 + err113 reference: docs/registry synced;
  the `v`/`h` receiver exceptions are REMOVED from the linter (decision:
  receiver is always the first letter); canonical error file is `error.go`
  (giderrfile default narrowed).
- §4 GID-233: `string(enum)` cast is allowed by design (plain-string targets);
  GID-215 narrowing documented as deliberate; GID-121 extended to all simple
  types except `*bool` (escape `//nolint:gidnoptr`).
- §5: implemented — GID-111 scope +`/client/**`; GID-121/123 scope +`/event/dto`;
  **GID-238/239 `gidcliflags`** (kebab-case Name, UPPER_SNAKE env, Required or
  default Value); **GID-240 `gidmodulealias`** + `pkg/<module>` module boundary
  in the layer matrix; constructor-returns-error softened in repository.md
  (may return error, not obliged — no rule needed).

Released as **v0.8.0**.
