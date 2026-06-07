// Eval для GID-194: settings.exclude отключает правило для перечисленных имён.
package excluded

const LegacyExported = 1 // исключена настройкой — диагностики нет

const OtherExported = 2 // want `GID-194: exported constant "OtherExported" is declared outside model/entity\. Fix: keep shared constants in /domain/model or /dal/entity, and declare local ones where they are used`
