// Eval for GID-194: settings.exclude disables the rule for the listed names.
package excluded

const LegacyExported = 1 // excluded by settings — no diagnostic

const OtherExported = 2 // want `GID-194: exported constant "OtherExported" is declared outside model/entity\. Fix: keep shared constants in /domain/model or /dal/entity, and declare local ones where they are used`
