# language: ru

Функция: GID-184 — сообщение ошибки описывает операцию, не факт провала (failedto)
  Как разработчик
  Я хочу, чтобы сообщение в errors.Wrap/Wrapf/WithMessage/WithMessagef/Errorf/New
  описывало выполняемую операцию ("select user"), а не факт провала ("failed to select user")
  Чтобы при разворачивании цепочки ошибок читалась последовательность операций

  # Один анализатор failedto → линтер gidfailedto, LoadModeTypesInfo.
  # pkg/errors распознаётся по import-пути github.com/pkg/errors через TypesInfo (стаб в testdata).
  # Проверяется только строковый литерал-сообщение; переменная/конкатенация с переменной — не матчатся.
  # Запрещённые префиксы (регистронезависимо, по границе слова):
  #   failed to, failed, unable to, error, couldn't, could not, can't, cannot
  # Список — дефолт, настраивается Settings{Prefixes []string `json:"prefixes"`} (замещает дефолт целиком).
  # Сгенерированный код (ast.IsGenerated) пропускается.

  # --- Класс 1: позитивный (нарушение ловится) ---

  Сценарий: позитивный — errors.Wrap с "failed to"
    Допустим "errors.Wrap(err, \"failed to select\")"
    Когда анализатор gidfailedto проверяет файл
    Тогда выводится диагностика "GID-184: сообщение ошибки начинается с \"failed to\" — опишите операцию: вместо \"failed to select user\" → \"select user\""

  Сценарий: позитивный — errors.New("Failed: x") в var (регистронезависимо)
    Допустим "var ErrSelect = errors.New(\"Failed: x\")"
    Когда анализатор gidfailedto проверяет файл
    Тогда выводится диагностика "GID-184: сообщение ошибки начинается с \"failed\""

  Сценарий: позитивный — errors.WithMessage с "unable to"
    Допустим "errors.WithMessage(err, \"unable to parse\")"
    Когда анализатор gidfailedto проверяет файл
    Тогда выводится диагностика "GID-184: сообщение ошибки начинается с \"unable to\""

  Сценарий: позитивный — errors.Errorf с "error"
    Допустим "errors.Errorf(\"error while loading %d\", id)"
    Когда анализатор gidfailedto проверяет файл
    Тогда выводится диагностика "GID-184: сообщение ошибки начинается с \"error\""

  Сценарий: позитивный — errors.Wrapf с "cannot" и errors.WithMessagef с "could not"
    Допустим "errors.Wrapf(err, \"cannot save %d\", id)" и "errors.WithMessagef(err, \"could not commit %d\", id)"
    Когда анализатор gidfailedto проверяет файл
    Тогда выводятся диагностики на оба вызова

  # --- Класс 2: негативный (чистый код проходит) ---

  Сценарий: негативный — сообщение описывает операцию
    Допустим "errors.Wrap(err, \"select user\")" и "errors.New(\"parse config\")"
    Когда анализатор gidfailedto проверяет файл
    Тогда диагностика не выводится

  # --- Класс 3: граничный (похоже на нарушение, но допустимо) ---

  Сценарий: граничный — "failure mode" (слово failure не в списке)
    Допустим "errors.Wrap(err, \"failure mode handling\")"
    Когда анализатор gidfailedto проверяет файл
    Тогда диагностика не выводится
    # "failure" не равно префиксу "failed"/"failed to"; граница слова защищает от подстрок.

  Сценарий: граничный — fmt.Sprintf не матчится (другой пакет)
    Допустим "errors.Wrap(err, fmt.Sprintf(\"%s\", \"x\"))"
    Когда анализатор gidfailedto проверяет файл
    Тогда диагностика не выводится
    # Аргумент-сообщение — не строковый литерал, а вызов fmt.Sprintf.

  Сценарий: граничный — std errors.New не матчится (зона GID-146)
    Допустим "stderrors.New(\"failed to do thing\")" из стандартного пакета "errors"
    Когда анализатор gidfailedto проверяет файл
    Тогда диагностика не выводится
    # Линтер матчит только github.com/pkg/errors; std errors — зона GID-146.

  Сценарий: граничный — не-литеральное сообщение (переменная / конкатенация)
    Допустим "errors.Wrap(err, msg)" и "errors.Wrap(err, \"failed to \"+name)"
    Когда анализатор gidfailedto проверяет файл
    Тогда диагностика не выводится
    # Переменная и конкатенация с переменной не имеют константного значения — не проверяются.

  # --- Класс 4: неприменимость ---

  Сценарий: неприменимость — файл без github.com/pkg/errors
    Допустим пакет с локальной функцией "Wrap(err, \"failed to select\")" без импорта pkg/errors
    Когда анализатор gidfailedto проверяет файл
    Тогда диагностика не выводится
    # Одноимённая локальная функция Wrap — не вызов pkg/errors (проверка через TypesInfo).

  # --- Класс 5: настройка ---

  Сценарий: настройка — settings.prefixes замещает дефолт целиком
    Допустим settings.prefixes = ["oops"], "errors.Wrap(err, \"oops broken\")" и "errors.Wrap(err, \"failed to select\")"
    Когда анализатор gidfailedto проверяет файл
    Тогда выводится диагностика только на "oops broken"; "failed to select" не ловится

# --- Чек-лист при добавлении нового правила ---
#  [x] ID и описание занесены в реестр (RULES.md, GID-184)
#  [x] Выбран слой: go/analysis (пакет failedto: gidfailedto), LoadModeTypesInfo
#  [x] Задано сообщение ("GID-184: …")
#  [x] Покрыты кейсы: позитивный, негативный, граничный, неприменимость, настройка
#  [x] testdata с // want для analysistest + стаб github.com/pkg/errors
#  [ ] Правило включено в .golangci.yml
