// Eval для GID-194: settings.exclude отключает правило для перечисленных имён.
package excluded

const LegacyExported = 1 // исключена настройкой — диагностики нет

const OtherExported = 2 // want `GID-194: экспортируемая константа "OtherExported" объявлена вне model/entity — общие константы живут в /domain/model или /dal/entity, локальные объявляются там, где используются`
