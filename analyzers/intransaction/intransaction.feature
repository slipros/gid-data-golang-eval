# language: en

Feature: GID-175 — the transaction-handling convention (in-transaction)
  As a developer
  I want the transaction type to live in /domain/model (InTransactionFunc),
  and a connection with this signature to be passed into the constructor directly
  So that service/usecase use a single named type instead of wrapping the transaction in methods

  # The canonical form in /domain/model:
  #   type InTransactionFunc func(ctx context.Context, fn func(ctx context.Context) error) error
  #   type InTransactionWithReturnFunc[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)
  #
  # Analyzer gidintransaction, LoadModeTypesInfo. The signature is matched structurally via go/types:
  #   plain:      params (context.Context, func(context.Context) error) -> error
  #   withReturn: params (context.Context, func(context.Context) (T, error)) -> (T, error)
  # context.Context is recognized by type (package context, name Context). Generated code is skipped.
  #
  # Checks:
  #   1. A tx-type declaration outside /domain/model.
  #   2. Naming of the tx type in /domain/model.
  #   3. An anonymous tx signature in /domain/service and /domain/usecase (field/parameter).
  #   4. A tx method on a struct in /dal/repository and /domain/service.

  # === Class 1: a tx-type declaration outside /domain/model (check 1) ===

  Scenario: positive — a named tx type declared outside model
    Given a package outside "/domain/model" (e.g. "/internal/pkg/helper") with the type "type Tx func(ctx context.Context, fn func(ctx context.Context) error) error"
    When the analyzer checks the file
    Then the diagnostic "GID-175: the transaction type must live in /domain/model (InTransactionFunc)" is reported on the type "Tx"

  Scenario: positive — a generic tx-type variant outside model
    Given a package outside "/domain/model" with the type "type TxR[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)"
    When the analyzer checks the file
    Then the diagnostic "GID-175: the transaction type must live in /domain/model (InTransactionFunc)" is reported on the type "TxR"

  # === Class 2: naming of the tx type in /domain/model (check 2) ===

  Scenario: positive — a tx type in model is not named InTransactionFunc
    Given a package in "/domain/model" with the type "type RunInTx func(ctx context.Context, fn func(ctx context.Context) error) error"
    When the analyzer checks the file
    Then the diagnostic "GID-175: the transaction type must be named InTransactionFunc / InTransactionWithReturnFunc" is reported on the type "RunInTx"

  Scenario: positive — a generic tx type in model is not named InTransactionWithReturnFunc
    Given a package in "/domain/model" with the type "type WithTxResult[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)"
    When the analyzer checks the file
    Then the diagnostic "GID-175: the transaction type must be named InTransactionFunc / InTransactionWithReturnFunc" is reported on the type "WithTxResult"

  # === Class 3: an anonymous tx signature in service/usecase (check 3) ===

  Scenario: positive — an anonymous tx signature in a service struct field
    Given a package in "/domain/service" with a struct having the field "tx func(ctx context.Context, fn func(ctx context.Context) error) error"
    When the analyzer checks the file
    Then the diagnostic "GID-175: use the named type model.InTransactionFunc" is reported on the field "tx"

  Scenario: positive — an anonymous tx signature in a constructor parameter
    Given a package in "/domain/service" with a constructor taking the parameter "tx func(ctx context.Context, fn func(ctx context.Context) error) error"
    When the analyzer checks the file
    Then the diagnostic "GID-175: use the named type model.InTransactionFunc" is reported on the parameter "tx"

  Scenario: positive — an anonymous generic tx signature in a usecase function parameter
    Given a package in "/domain/usecase" with a function taking the parameter "run func(ctx context.Context, fn func(ctx context.Context) (string, error)) (string, error)"
    When the analyzer checks the file
    Then the diagnostic "GID-175: use the named type model.InTransactionFunc" is reported on the parameter "run"

  # === Class 4: a tx method on repo/service (check 4) ===

  Scenario: positive — a tx method on a repository (any name)
    Given a package in "/dal/repository" with a struct and the method "func (r *JobRepository) InTx(ctx context.Context, fn func(ctx context.Context) error) error"
    When the analyzer checks the file
    Then the diagnostic "GID-175: a repository/service must not wrap a transaction in a method — InTransactionFunc is passed into the constructor directly from the connection" is reported on the method "InTx"

  Scenario: positive — a tx method on a service (any name)
    Given a package in "/domain/service" with a struct and the method "func (s *JobService) Transaction(ctx context.Context, fn func(ctx context.Context) error) error"
    When the analyzer checks the file
    Then the diagnostic "GID-175: a repository/service must not wrap a transaction in a method" is reported on the method "Transaction"

  # === Negative cases (clean code) ===

  Scenario: negative — the canonical model InTransactionFunc / InTransactionWithReturnFunc
    Given a package in "/domain/model" with the types "InTransactionFunc" and "InTransactionWithReturnFunc[T any]" of the canonical form
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a service with a field of the named type model.InTransactionFunc
    Given a package in "/domain/service" with a struct having the field "tx model.InTransactionFunc" and a constructor taking "tx model.InTransactionFunc"
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Boundary cases (a similar but different signature — not flagged) ===

  Scenario: boundary — a callback with an extra argument
    Given the type "func(ctx context.Context, fn func(ctx context.Context, id int) error) error"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — without ctx as the first parameter
    Given the type "func(fn func(ctx context.Context) error) error"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — the callback does not return error
    Given the method "func (r *JobRepository) NotInTx(ctx context.Context, fn func(ctx context.Context) (int, error)) error"
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Non-applicability ===

  Scenario: non-applicability — an anonymous tx signature outside service/usecase (in main)
    Given the package "main" with a function taking the parameter "tx func(ctx context.Context, fn func(ctx context.Context) error) error"
    When the analyzer checks the file
    Then no diagnostic is reported
    # (Check 3 applies only in /domain/service and /domain/usecase.
    #  Check 1 catches only named type declarations, not anonymous parameters.)

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-175)
#  [x] Layer chosen: go/analysis (analyzer gidintransaction in analyzers/intransaction)
#  [x] Severity and message are defined ("GID-175: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
