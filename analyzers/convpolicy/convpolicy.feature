# language: en

Feature: GID-247 — convert-no-policy: a converter maps, it does not decide
  As a service architect
  I want convert functions not to branch on their input to pick a raw constant value
  So that domain policy (codec, sample rate, channel count) stays in /domain/model
    and the converter remains a pure input → output mapping

  Scenario: if on an input parameter selects a raw channel count — violation
    Given the convert function "AudioFormatFromSource" initializes "channels" with a basic constant
    And reassigns it to a different basic constant inside "if source == SourceMeeting"
    When the analyzer checks the file
    Then a "GID-247" diagnostic is reported on the in-branch assignment to "channels"

  Scenario: switch on an input parameter selects a raw sample rate — violation
    Given the convert function "SampleRateFromSource" assigns "rate" two different basic constants
    And the assignments sit in case clauses of "switch source"
    When the analyzer checks the file
    Then a "GID-247" diagnostic is reported on the first in-branch assignment to "rate"

  Scenario: the converter copies fields from its input — ok
    Given the convert function "AudioFormatCopy" assigns output fields from "in"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: the branch selects among named enum constants (Codec) — ok
    Given the convert function "CodecFromSource" switches on input and assigns "c" the named enum values CodecOpus/CodecAAC
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: the branch condition tests a local value, not an input parameter — ok
    Given the convert function "LocalBranch" branches on a locally computed "n", not on the parameter
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: every branch assigns the same constant — a single distinct value, not a selection — ok
    Given the convert function "SameConst" assigns "x" the value 5 in both branches of "if source == SourceMeeting"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: the same policy pattern in a non-convert package — the rule does not apply
    Given the function "pickChannels" lives in package "svc/domain/service" (not a convert package)
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: settings.exclude exempts a named converter
    Given the analyzer is configured with settings.exclude = ["asrFormatFromSource"]
    And the convert package "exsvc/domain/service/convert" contains "asrFormatFromSource" and "SampleRateFromSource"
    When the analyzer checks the file
    Then no diagnostic is reported on "asrFormatFromSource"
    And a "GID-247" diagnostic is reported on the in-branch assignment to "rate" in "SampleRateFromSource"

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the PRD registry (RULES.md)
#  [x] Layer chosen: go/analysis (types info is needed — constant values and parameter objects)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
