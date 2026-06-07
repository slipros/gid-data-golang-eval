# language: ru

Функция: GID-215 — конвертация model ↔ entity живёт только в convert-пакетах (no-inline-entity-literal)
  Как разработчик
  Я хочу, чтобы инлайн-заполнение entity-типов в domain-слое было запрещено
  Чтобы вся конвертация model ↔ entity жила в пакете convert (<Dst><Type>From<Src>)

  # Один анализатор inlineconv → линтер gidinlineconv, LoadModeTypesInfo.
  # Источник: service.md «Конвертация всегда выполняется через пакет convert».
  # Scope: пакеты domain-слоя (pathseg.Contains(pkgPath, "domain")), КРОМЕ пакетов
  # с сегментом convert. Тип литерала определяется через TypesInfo: именованный
  # тип (struct или именованный слайс) из пакета entity-слоя
  # (pathseg.Contains(пакет типа, "dal", "entity"), включая filter/enum).
  # Пустой литерал (entity.Snapshot{}) — zero value, разрешён.
  # Флагается только внешний (outermost) entity-литерал — внутрь не спускаемся.
  # _test.go и сгенерированный код (ast.IsGenerated) пропускаются.

  # --- Класс 1: позитивный (нарушение ловится) ---

  Сценарий: позитивный — entity.CreateSnapshot{Field: ...} в /domain/service
    Допустим "return entity.CreateSnapshot{Name: name}" в пакете /domain/service
    Когда анализатор gidinlineconv проверяет файл
    Тогда выводится диагностика "GID-215: инлайн-заполнение entity-типа entity.CreateSnapshot в domain-слое запрещено — конвертация живёт в convert-пакете (<Dst><Type>From<Src>)"

  Сценарий: позитивный — &entity.Snapshot{Field: ...}
    Допустим "return &entity.Snapshot{ID: id}" в пакете /domain/service
    Когда анализатор gidinlineconv проверяет файл
    Тогда выводится диагностика "GID-215: инлайн-заполнение entity-типа entity.Snapshot в domain-слое запрещено"

  Сценарий: позитивный — именованный entity-слайс с элементами
    Допустим "return entity.Snapshots{{ID: \"a\"}, {ID: \"b\"}}" в пакете /domain/service
    Когда анализатор gidinlineconv проверяет файл
    Тогда выводится диагностика "GID-215: инлайн-заполнение entity-типа entity.Snapshots в domain-слое запрещено"

  Сценарий: позитивный — filter-структура из /dal/entity/filter с полями
    Допустим "return filter.Snapshots{Name: name, Limit: 10}" в пакете /domain/service
    Когда анализатор gidinlineconv проверяет файл
    Тогда выводится диагностика "GID-215: инлайн-заполнение entity-типа filter.Snapshots в domain-слое запрещено"

  # --- Класс 2: негативный (чистый код проходит) ---

  Сценарий: негативный — пустой entity-литерал (zero value)
    Допустим "return entity.Snapshot{}" в пакете /domain/service
    Когда анализатор gidinlineconv проверяет файл
    Тогда диагностика не выводится
    # Пустой литерал — zero value, не инлайн-конвертация.

  Сценарий: негативный — литерал model-типа с полями
    Допустим "return model.Snapshot{ID: id, Name: name}" в пакете /domain/service
    Когда анализатор gidinlineconv проверяет файл
    Тогда диагностика не выводится
    # model в domain — норма; правило только про entity-типы.

  Сценарий: негативный — entity-литерал внутри convert-пакета сервиса
    Допустим "return entity.CreateSnapshot{Name: in.Name}" в пакете /domain/service/convert
    Когда анализатор gidinlineconv проверяет файл
    Тогда диагностика не выводится
    # convert-пакет — место конвертации, инлайн entity разрешён.

  # --- Класс 3: граничный ---

  Сценарий: граничный — вложенный литерал внутри зафлаганного внешнего
    Допустим "entity.Snapshots{entity.Snapshot{ID: \"a\"}, entity.Snapshot{ID: \"b\"}}" в пакете /domain/service
    Когда анализатор gidinlineconv проверяет файл
    Тогда выводится ровно одна диагностика — на внешний литерал entity.Snapshots
    # Внутрь зафлаганного литерала анализатор не спускается.

  Сценарий: граничный — map[string]entity.X{ key: entity.X{...} }
    Допустим "map[string]entity.Snapshot{id: entity.Snapshot{ID: id}}" в пакете /domain/service
    Когда анализатор gidinlineconv проверяет файл
    Тогда выводится диагностика на значение entity.Snapshot, но не на map-литерал
    # Сам map-литерал — не именованный entity-тип (не флагается); флагается
    # значение entity.Snapshot, в т.ч. элемент без явного типа ({ID: id}).

  # --- Класс 4: неприменимость ---

  Сценарий: неприменимость — entity-литерал в /dal/repository
    Допустим "return entity.Snapshot{ID: id}" в пакете /dal/repository
    Когда анализатор gidinlineconv проверяет файл
    Тогда диагностика не выводится
    # /dal/repository не входит в domain-слой — не зона правила.

  Сценарий: неприменимость — _test.go
    Допустим "return entity.Snapshot{ID: id}" в файле service_test.go
    Когда анализатор gidinlineconv проверяет файл
    Тогда диагностика не выводится
    # Тестовые файлы пропускаются.

# --- Чек-лист при добавлении нового правила ---
#  [x] ID и описание занесены в реестр (RULES.md, GID-215)
#  [x] Выбран слой: go/analysis (пакет inlineconv: gidinlineconv), LoadModeTypesInfo
#  [x] Задано сообщение ("GID-215: …")
#  [x] Покрыты кейсы: позитивный, негативный, граничный, неприменимость
#  [x] testdata с // want для analysistest
#  [ ] Правило включено в .golangci.yml
