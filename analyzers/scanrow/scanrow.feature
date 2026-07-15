# language: en

Feature: GID-245 — an epgx Select into a one-field struct should use ScanRow (scanrow)
  As a developer
  I want a single-column read to use conn.ScanRow with the column pointer, not conn.Select into a one-field struct
  So that a single-value read is expressed directly instead of mapping one column into a struct

  # One analyzer, linter gidscanrow, LoadModeTypesInfo.
  # epgx (git.k8s.nomilk.space/go-library/epgx): Connection.Select(ctx, ptr any, sql, args...)
  # maps columns into a struct/slice; Connection.ScanRow(ctx, scan []any, sql, args...)
  # assigns a single row's columns into a slice of scalar pointers.
  #
  # Match (deliberately narrow, to avoid false positives — all required):
  #   - a call x.Select(ctx, arg, ...);
  #   - the receiver type of Select also exposes a method
  #     ScanRow(context.Context, []any, string, ...any) error — the epgx
  #     fingerprint (its second parameter is a slice), so a foreign .Select is
  #     never flagged;
  #   - arg's type is a pointer to a struct with EXACTLY one field.
  # Slices (*[]T, the multi-row form) and multi-field structs are out of scope.
  # Fix: var out T; conn.ScanRow(ctx, []any{&out}, sql, args...).

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — Select into an anonymous one-field struct
    Given "var out struct { MemberID string `db:\"member_id\"` }; conn.Select(ctx, &out, sql)" where conn also has ScanRow
    When the gidscanrow analyzer checks the file
    Then the diagnostic "GID-245: Select into a single-field struct reads one column — use ScanRow with the field pointer …" is reported on the call

  Scenario: positive — Select into a named one-field struct
    Given "type oneCol struct { ID string }; var out oneCol; conn.Select(ctx, &out, sql)"
    When the gidscanrow analyzer checks the file
    Then the diagnostic "GID-245 …" is reported on the call

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — Select into a multi-field struct
    Given "type twoCol struct { ID, Name string }; var out twoCol; conn.Select(ctx, &out, sql)"
    When the gidscanrow analyzer checks the file
    Then no diagnostic is reported
    # More than one column — Select is the right call.

  Scenario: negative — Select into a slice (the multi-row form)
    Given "var out []oneCol; conn.Select(ctx, &out, sql)"
    When the gidscanrow analyzer checks the file
    Then no diagnostic is reported
    # *[]T is not a struct pointee — reading many rows, ScanRow does not apply.

  Scenario: negative — already ScanRow
    Given "var out string; conn.ScanRow(ctx, []any{&out}, sql)"
    When the gidscanrow analyzer checks the file
    Then no diagnostic is reported
    # The method is ScanRow, not Select.

  # --- Class 3: boundary (looks like a violation but is allowed) ---

  Scenario: boundary — Select into a one-field struct on a non-epgx receiver
    Given a receiver that has Select but NO ScanRow method, and "var out oneCol; other.Select(ctx, &out, sql)"
    When the gidscanrow analyzer checks the file
    Then no diagnostic is reported
    # Without the ScanRow fingerprint the receiver is not confirmed to be an epgx connection.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — settings.exclude exempts a specific method
    Given settings.exclude contains "Repo.excludedMethod"
    When the gidscanrow analyzer checks that method
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-245)
#  [x] Layer chosen: go/analysis (package scanrow: gidscanrow)
#  [x] Message is defined ("GID-245: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
