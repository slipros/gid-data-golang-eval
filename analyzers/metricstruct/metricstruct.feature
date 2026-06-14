# language: en
# Specification of rule GID-174 (metric-prometheus-struct).
# Linter: gidmetricstruct (go/analysis, LoadModeTypesInfo).

Feature: GID-174 — the standardized service metrics package
  As a backend-go service developer
  I want the metrics package to follow a single convention
  So that in all services metrics live under the /metric path, in the metric package,
  and are aggregated in the exported Prometheus struct with a Register method

  # The grouping convention: extra metrics live in SEPARATE files, grouped
  # functionally in structs (one group per file). prometheus.go is the
  # wiring: the Prometheus type is declared there, its Register calls the groups' Register.

  # --- Positive cases (the violation is caught) ---

  Scenario: the package path ends with metrics — violation
    Given a package whose import path ends with the segment "metrics"
    When the analyzer checks the package
    Then the diagnostic "GID-174: the metrics package must be named metric, not metrics" is reported on the package clause

  Scenario: the path ends with metric but there is no Prometheus type — violation
    Given a package at the path ".../metric" without a Prometheus type
    When the analyzer checks the package
    Then the diagnostic "GID-174: the metric package must declare a metrics aggregator: struct Prometheus with a Register method" is reported
    And the diagnostic is reported on the package clause of the file with the smallest name (deterministic)

  Scenario: Prometheus exists but has no Register method — violation
    Given a package at the path ".../metric" with a Prometheus struct type without a Register method
    When the analyzer checks the package
    Then a diagnostic about the missing Register method is reported on the Prometheus type declaration

  Scenario: Prometheus is declared but is not a struct — violation
    Given a package at the path ".../metric" with a Prometheus type that is not a struct
    When the analyzer checks the package
    Then a diagnostic saying Prometheus must be a struct is reported on the type declaration

  Scenario: Prometheus is declared outside prometheus.go — violation
    Given a package at the path ".../metric" with the Prometheus struct in the file metric.go
    When the analyzer checks the package
    Then the diagnostic "GID-174: the Prometheus aggregator must live in prometheus.go" is reported on the type declaration

  Scenario: another exported struct group is declared in prometheus.go — violation
    Given prometheus.go contains Prometheus and the exported struct type HTTPMetrics
    When the analyzer checks the package
    Then the diagnostic "GID-174: a metrics group must live in its own file; prometheus.go is wiring only" is reported on the HTTPMetrics declaration

  Scenario: two functional groups in one file — violation
    Given one file of the package (not prometheus.go) declares ≥2 exported struct types
    When the analyzer checks the package
    Then the diagnostic "GID-174: one functional metrics group per file" is reported on the second and subsequent ones

  Scenario: a group field is not registered in Prometheus.Register — violation
    Given a Prometheus field has a type with a Register method, but the body of Prometheus.Register has no call to its Register
    When the analyzer checks the package
    Then the diagnostic "GID-174: Prometheus.Register registers group %s — call its Register" is reported on the field

  Scenario: an embedded group field is not registered — violation
    Given Prometheus embeds a type with a Register method without calling its Register in Prometheus.Register
    When the analyzer checks the package
    Then a diagnostic about the need to call the embedded group's Register is reported

  # --- Negative cases (clean code passes) ---

  Scenario: a canonical metric package with wiring and groups — ok
    Given prometheus.go declares Prometheus, the groups are in separate files,
      and Prometheus.Register calls the Register of every group field
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: Register with a pointer receiver — ok
    Given the Register method is declared on a pointer (*Prometheus)
    When the analyzer checks the package
    Then no diagnostic is reported

  # --- Boundary cases ---

  Scenario: Register with any signature (with parameters and a return value) — ok
    Given the Register method has an arbitrary signature
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: a generated file is not chosen for the package-clause report
    Given the metric package contains a generated file and an ordinary file
    When the analyzer reports on the package clause
    Then the report is placed on the non-generated file

  Scenario: a group field without a Register method does not require registration — ok
    Given a Prometheus field has a type without a Register method (e.g. int)
    When the analyzer checks the package
    Then no diagnostic is reported on this field

  # --- Non-applicability (the rule does not apply) ---

  Scenario: the package path does not end with metric/metrics — the rule does not apply
    Given a package outside a metric path, even if it has a Prometheus type without Register
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: a main or auxiliary package — the scope is by path only
    Given a package whose path does not end with metric/metrics
    When the analyzer checks the package
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-174)
#  [x] Layer chosen: go/analysis (types needed — the Register method via types)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
