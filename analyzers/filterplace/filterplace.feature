# language: en

Feature: GID-171 — filters of list operations live in their layer's designated place
  As a developer
  I want filter structs to live in the dedicated place of their layer
  So that entity filters and model filters do not sprawl across repository/service
  Sources: model.md, entity.md
  Scope: packages in /dal/** and /domain/**
  A STRUCT type declaration is checked (struct only — to avoid flagging
  FilterFunc, interfaces and aliases) with a filter name.
  A name counts as a filter if the word Filter stands at a boundary:
    - the prefix Filter + an uppercase letter/digit or the end of the name (Filter, FilterJobs);
    - the suffix Filter after a lowercase letter/digit or at the start (JobsFilter).
  Filterable is NOT flagged: Filter is followed by a lowercase letter — it is a different word.

  # --- Positive class: the violation is caught ---

  Scenario: a *Filter struct in /dal/repository — violation
    Given "type JobsFilter struct{...}" is declared in /dal/repository
    When the analyzer checks the file
    Then a "GID-171" diagnostic is reported with the text "must live in /dal/entity/filter"

  Scenario: a Filter* struct in /domain/service — violation
    Given "type FilterJobs struct{...}" is declared in /domain/service
    When the analyzer checks the file
    Then a "GID-171" diagnostic is reported with the text "must live in /domain/model"

  # --- Negative class: clean code passes ---

  Scenario: an entity filter in /dal/entity/filter — ok
    Given "type JobsFilter struct{...}" is declared in /dal/entity/filter
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a model filter in /domain/model — ok
    Given "type JobsFilter struct{...}" is declared in /domain/model
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a model filter in the /domain/model/filter subpackage — ok
    Given "type JobsFilter struct{...}" is declared in /domain/model/filter
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary class ---

  Scenario: FilterFunc (a func type) in /dal/repository — not a struct, not flagged
    Given "type FilterFunc func(string) bool" is declared in /dal/repository
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a Filterable struct in /dal/repository — the name does not match the pattern, ok
    Given "type Filterable struct{...}" is declared in /dal/repository
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a generated file — skipped
    Given the file is marked "// Code generated ... DO NOT EDIT."
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Non-applicability class: the rule does not apply outside dal/domain ---

  Scenario: a *Filter struct in /server/http — the rule does not apply
    Given "type JobsFilter struct{...}" is declared in /server/http
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-171)
#  [x] Layer chosen: go/analysis (pure AST, LoadModeSyntax)
#  [x] Severity and message are defined ("GID-171: ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (outside the scope of this task)
