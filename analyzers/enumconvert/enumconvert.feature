# language: ru

Функция: GID-143 — обработка отсутствующего ключа enum-конвертации (enumconvert)
  Как разработчик
  Я хочу, чтобы map-конвертация enum обрабатывала отсутствующий ключ
  через gderror.NewUnhandledValueError
  Чтобы неизвестное значение enum не превращалось молча в zero-value

  # Анализатор enumconvert, линтер gidenumconvert, LoadMode TypesInfo.
  # Scope: только convert-пакеты (последний сегмент пути — convert),
  #   матчится через internal/pathseg.EndsWith.
  # Детект (по типам): индексация мапы m[key], где тип ключа — именованный
  #   тип с underlying string (enum по GID-123), а тип значения — тоже
  #   именованный тип (enum→enum / enum→модельный тип).
  # gderror распознаётся по import-пути
  #   gitlab.gid.team/gid-data/tech/golang/libs/helper.git/errors
  #   и имени конструктора NewUnhandledValueError (стаб в testdata).
  # Сгенерированный код (ast.IsGenerated) пропускается.

  # --- Класс 1: позитивный (нарушение ловится) ---

  Сценарий: позитивный — индексация без comma-ok (одиночное присваивание)
    Допустим convert-пакет с "v := statusMap[s]", где ключ — enum, значение — именованный тип
    Когда анализатор gidenumconvert проверяет файл
    Тогда выводится диагностика "GID-143: enum-конвертация через map без comma-ok — отсутствующий ключ должен давать gderror.NewUnhandledValueError" на "statusMap[s]"

  Сценарий: позитивный — индексация без comma-ok (использование выражением)
    Допустим convert-пакет с "return statusMap[s]"
    Когда анализатор gidenumconvert проверяет файл
    Тогда выводится диагностика "GID-143: enum-конвертация через map без comma-ok …"

  Сценарий: позитивный — comma-ok есть, но в функции нет NewUnhandledValueError
    Допустим convert-пакет с "v, ok := statusMap[s]" и без вызова gderror.NewUnhandledValueError в теле функции
    Когда анализатор gidenumconvert проверяет файл
    Тогда выводится диагностика "GID-143: отсутствующий ключ enum-конвертации обрабатывается gderror.NewUnhandledValueError" на "statusMap[s]"

  # --- Класс 2: негативный (чистый код проходит) ---

  Сценарий: негативный — comma-ok + обработка отсутствующего ключа
    Допустим convert-пакет с "v, ok := statusMap[s]; if !ok { return \"\", gderror.NewUnhandledValueError(s) }"
    Когда анализатор gidenumconvert проверяет файл
    Тогда диагностика не выводится

  # --- Класс 3: граничный (похоже на нарушение, но допустимо) ---

  Сценарий: граничный — ключ мапы базовый string (не enum)
    Допустим convert-пакет с "titleMap[s]", где ключ — string, а не именованный enum
    Когда анализатор gidenumconvert проверяет файл
    Тогда диагностика не выводится
    # Базовый ключ string/int — не enum-конвертация.

  Сценарий: граничный — значение мапы базовый тип (не именованный)
    Допустим convert-пакет с "weightMap[s]", где значение — int
    Когда анализатор gidenumconvert проверяет файл
    Тогда диагностика не выводится
    # Значение не именованный тип — это не enum→enum/модель.

  Сценарий: граничный — та же конструкция вне convert-пакета
    Допустим пакет в "/domain/service" (не convert) с "return statusMap[s]"
    Когда анализатор gidenumconvert проверяет файл
    Тогда диагностика не выводится
    # Scope — только convert-пакеты.

  # --- Класс 4: неприменимость ---

  Сценарий: неприменимость — convert-пакет без map-индексаций enum
    Допустим convert-пакет с обычными полевыми конвертерами без мап
    Когда анализатор gidenumconvert проверяет файл
    Тогда диагностика не выводится

# --- Чек-лист при добавлении нового правила ---
#  [x] ID и описание занесены в реестр (RULES.md, GID-143)
#  [x] Выбран слой: go/analysis (пакет enumconvert: gidenumconvert)
#  [x] Заданы сообщения ("GID-143: …")
#  [x] Покрыты кейсы: позитивный, негативный, граничный, неприменимость
#  [x] testdata с // want для analysistest
#  [ ] Правило включено в .golangci.yml
