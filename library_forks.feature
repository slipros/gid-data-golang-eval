Feature: Library import path overrides
  Every custom linter accepts a per-linter libraries mapping.
  Mappings match import path segment prefixes and preserve subpackages.
  Configured analyzers retain independent immutable settings.

  Scenario: A fork provides the corporate PostgreSQL API
    Given gidbansymbol maps the corporate postgres library to fork.example/postgres
    When a call resolves to fork.example/postgres.TQuery
    Then GID-217 reports the banned symbol

  Scenario: Recommendations point to the configured helper fork
    Given gidallptr maps the corporate helper library to fork.example/helper
    When a slice of structs is iterated by value
    Then the recommendation refers to fork.example/helper

  Scenario: Prefix boundaries prevent unrelated package matches
    Given corp.example/lib is replaced by fork.example/lib
    When a package imports fork.example/library
    Then its path remains unrelated to corp.example/lib

  Scenario: An unconfigured analyzer retains its defaults
    Given another analyzer has a library override
    When an analyzer without libraries runs concurrently
    Then its package matching and diagnostics use their default paths

  Scenario: Ambiguous or invalid settings fail at configuration time
    Given multiple source libraries map to the same fork or a path is invalid
    When the plugin is configured
    Then configuration fails with an explicit error
