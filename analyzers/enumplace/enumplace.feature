# language: ru
# Eval-спека правила GID-211 (enum-location), линтер gidenumplace.

Функция: GID-211 — enum DAL-слоя живут в /dal/entity/enum
  Как разработчик DAL-слоя
  Я хочу, чтобы string-enum объявлялись только в /dal/entity/enum
  Чтобы каждый enum лежал в отдельном файле по имени сущности (entity.md)

  Сценарий: string-enum в /dal/entity (корне) — нарушение
    Допустим в пакете /dal/entity объявлен "type Status string" с const этого типа
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-211" на имени типа "Status"

  Сценарий: string-enum в /dal/repository — нарушение
    Допустим в пакете /dal/repository объявлен "type Mode string" с const этого типа
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-211" на имени типа "Mode"

  Сценарий: string-enum в /dal/entity/enum — ок
    Допустим в пакете /dal/entity/enum объявлен "type Status string" с const этого типа
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: string-тип без const в DAL — не enum
    Допустим в пакете /dal/entity объявлен "type RawJSON string" без const этого типа
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: alias на string с const — зона GID-123, не GID-211
    Допустим в пакете /dal/entity объявлен "type Code = string" с const
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: именованный int-тип с const — не string-enum
    Допустим в пакете /dal/entity объявлен "type Priority int" с const этого типа
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: string-enum в /domain/model — правило не применяется
    Допустим в пакете /domain/model объявлен "type Status string" с const этого типа
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

# --- Чек-лист при добавлении нового правила ---
#  [x] ID и описание занесены в реестр (RULES.md)
#  [x] Выбран слой: go/analysis (нужна информация о типах)
#  [x] Заданы severity и сообщение
#  [x] Покрыты кейсы: позитивный, негативный, граничный, неприменимость
#  [x] testdata с // want для analysistest
#  [x] Правило включено в .golangci.yml
