Feature: Explicit repository and orchestration code shapes
  Production code uses named intermediates and exposes failures before results.
  Generated files and _test.go fixtures are outside these policies.
  Thresholds count semantic items, regardless of source line width.

  Scenario Outline: Positive violations
    Given production code in the applicable layer contains <shape>
    When gidcodepolicy analyzes the package with Go type information
    Then it reports <rule> with an actionable example
    Examples:
      | rule    | shape                                                         |
      | GID-281 | 4 map entries, 5 struct fields or 8 call/signature parameters packed together |
      | GID-282 | a type alias in dal/repository                                |
      | GID-283 | a single-line SQL constant longer than 80 characters           |
      | GID-284 | repository SQL with an executable $1 placeholder               |
      | GID-285 | a gdpostgres.NamedArgs literal supplied directly to a call     |
      | GID-286 | a local empty-struct map used exclusively for membership lookup and insertion |
      | GID-287 | a temporary composite literal used as the receiver of a resolved method          |
      | GID-288 | return created, err immediately after its terminating error guard |
      | GID-289 | fields of a failed convert-package result used in its error branch |
      | GID-290 | a service/usecase loop appending individually converted values |
      | GID-291 | typed enum dispatch in an error-returning function without an unhandled-value default |
      | GID-292 | service/usecase json.Marshal of an anonymous struct            |
      | GID-293 | a nonempty string literal operand of == or !=                  |
      | GID-294 | storing a conversion error in a field and falling through      |
      | GID-295 | adjacent function declarations without a blank line            |
      | GID-296 | adjacent package-level error variables in separate var declarations |

  Scenario Outline: Negative forms
    Given production code uses <shape>
    When gidcodepolicy analyzes the package
    Then it emits no diagnostic for <rule>
    Examples:
      | rule    | shape                                                         |
      | GID-281 | one item per line and a separate closing delimiter             |
      | GID-282 | the entity type directly or a distinct defined type            |
      | GID-283 | a multiline SQL constant                                      |
      | GID-284 | named SQL placeholders                                        |
      | GID-285 | a separately declared args variable                            |
      | GID-286 | a ready helper set                                            |
      | GID-287 | a named variable as the method receiver                      |
      | GID-288 | return created, nil after the guard                            |
      | GID-289 | returning an error before accessing the converted value        |
      | GID-290 | a batch conversion call                                       |
      | GID-291 | an enum switch default returning gderr.NewUnhandledValueError  |
      | GID-292 | a named domain/model type with json tags                       |
      | GID-293 | a named constant as the comparison operand                     |
      | GID-294 | propagating the conversion error                               |
      | GID-295 | a blank line between functions                                |
      | GID-296 | a single var block for adjacent sentinels                      |

  Scenario Outline: Boundaries
    Given production code contains <shape>
    When gidcodepolicy analyzes the package
    Then it emits no diagnostic for <rule>
    Examples:
      | rule    | shape                                                         |
      | GID-281 | 3 map entries, 4 struct fields or 7 parameters                  |
      | GID-282 | a defined repository type without an equals token              |
      | GID-283 | SELECT 1                                                      |
      | GID-284 | $1 inside SQL strings, quoted identifiers or comments          |
      | GID-285 | a different library type coincidentally named NamedArgs        |
      | GID-286 | an empty-struct map returned to a caller or passed elsewhere    |
      | GID-287 | a named variable with a literal initializer or a function-valued field                     |
      | GID-288 | an error modified between its guard and the return             |
      | GID-289 | a successful conversion result used after its error guard      |
      | GID-290 | conversion whose result is not appended to a collection        |
      | GID-291 | a one-branch enum filter without else                          |
      | GID-292 | a named struct type                                           |
      | GID-293 | comparison with the empty string, numeric zero or a boolean     |
      | GID-294 | an error branch that returns                                   |
      | GID-295 | a declaration separating two functions                         |
      | GID-296 | non-error variables separating sentinel declarations           |

  Scenario: Non-applicability
    Given generated code or test scaffolding with these shapes
    Or repository/domain-specific shapes in an unrelated package
    When gidcodepolicy analyzes the package
    Then it skips generated files and test files
    And applies layer-specific checks only to their declared layers
    And general layout, temporary-method, success-return, set and literal checks remain applicable to production packages

  Scenario: A batch converter performs its own per-element conversion
    Given a package under domain/service/convert or domain/usecase/convert
    When a batch function converts each input and appends its output
    Then GID-290 is not reported because conversion belongs in that package
