# language: ru

Функция: GID-175 — конвенция работы с транзакциями (in-transaction)
  Как разработчик
  Я хочу, чтобы тип транзакции жил в /domain/model (InTransactionFunc),
  а connection с этой сигнатурой передавался в конструктор напрямую
  Чтобы service/usecase использовали единый именованный тип, а не оборачивали транзакцию методами

  # Каноническая форма в /domain/model:
  #   type InTransactionFunc func(ctx context.Context, fn func(ctx context.Context) error) error
  #   type InTransactionWithReturnFunc[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)
  #
  # Анализатор gidintransaction, LoadModeTypesInfo. Сигнатура матчится структурно через go/types:
  #   plain:      params (context.Context, func(context.Context) error) -> error
  #   withReturn: params (context.Context, func(context.Context) (T, error)) -> (T, error)
  # context.Context распознаётся по типу (пакет context, имя Context). Сгенерированный код пропускается.
  #
  # Проверки:
  #   1. Объявление tx-типа вне /domain/model.
  #   2. Нейминг tx-типа в /domain/model.
  #   3. Анонимная tx-сигнатура в /domain/service и /domain/usecase (поле/параметр).
  #   4. Tx-метод структуры в /dal/repository и /domain/service.

  # === Класс 1: объявление tx-типа вне /domain/model (проверка 1) ===

  Сценарий: позитивный — именованный tx-тип объявлен вне model
    Допустим пакет вне "/domain/model" (например "/internal/pkg/helper") с типом "type Tx func(ctx context.Context, fn func(ctx context.Context) error) error"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-175: тип транзакции живёт в /domain/model (InTransactionFunc)" на типе "Tx"

  Сценарий: позитивный — generic-вариант tx-типа вне model
    Допустим пакет вне "/domain/model" с типом "type TxR[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-175: тип транзакции живёт в /domain/model (InTransactionFunc)" на типе "TxR"

  # === Класс 2: нейминг tx-типа в /domain/model (проверка 2) ===

  Сценарий: позитивный — tx-тип в model назван не InTransactionFunc
    Допустим пакет в "/domain/model" с типом "type RunInTx func(ctx context.Context, fn func(ctx context.Context) error) error"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-175: тип транзакции называется InTransactionFunc / InTransactionWithReturnFunc" на типе "RunInTx"

  Сценарий: позитивный — generic-tx-тип в model назван не InTransactionWithReturnFunc
    Допустим пакет в "/domain/model" с типом "type WithTxResult[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-175: тип транзакции называется InTransactionFunc / InTransactionWithReturnFunc" на типе "WithTxResult"

  # === Класс 3: анонимная tx-сигнатура в service/usecase (проверка 3) ===

  Сценарий: позитивный — анонимная tx-сигнатура в поле структуры сервиса
    Допустим пакет в "/domain/service" со структурой с полем "tx func(ctx context.Context, fn func(ctx context.Context) error) error"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-175: используйте именованный тип model.InTransactionFunc" на поле "tx"

  Сценарий: позитивный — анонимная tx-сигнатура в параметре конструктора
    Допустим пакет в "/domain/service" с конструктором, принимающим параметр "tx func(ctx context.Context, fn func(ctx context.Context) error) error"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-175: используйте именованный тип model.InTransactionFunc" на параметре "tx"

  Сценарий: позитивный — анонимная generic-tx-сигнатура в параметре функции usecase
    Допустим пакет в "/domain/usecase" с функцией, принимающей параметр "run func(ctx context.Context, fn func(ctx context.Context) (string, error)) (string, error)"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-175: используйте именованный тип model.InTransactionFunc" на параметре "run"

  # === Класс 4: tx-метод на repo/service (проверка 4) ===

  Сценарий: позитивный — tx-метод на репозитории (имя любое)
    Допустим пакет в "/dal/repository" со структурой и методом "func (r *JobRepository) InTx(ctx context.Context, fn func(ctx context.Context) error) error"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-175: репозиторий/сервис не оборачивает транзакцию методом — InTransactionFunc передаётся в конструктор напрямую от connection" на методе "InTx"

  Сценарий: позитивный — tx-метод на сервисе (имя любое)
    Допустим пакет в "/domain/service" со структурой и методом "func (s *JobService) Transaction(ctx context.Context, fn func(ctx context.Context) error) error"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-175: репозиторий/сервис не оборачивает транзакцию методом" на методе "Transaction"

  # === Негативные кейсы (чистый код) ===

  Сценарий: негативный — каноническая модель InTransactionFunc / InTransactionWithReturnFunc
    Допустим пакет в "/domain/model" с типами "InTransactionFunc" и "InTransactionWithReturnFunc[T any]" канонической формы
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: негативный — сервис с полем именованного типа model.InTransactionFunc
    Допустим пакет в "/domain/service" со структурой с полем "tx model.InTransactionFunc" и конструктором, принимающим "tx model.InTransactionFunc"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  # === Граничные кейсы (похожая, но другая сигнатура — не флагуем) ===

  Сценарий: граничный — callback с дополнительным аргументом
    Допустим тип "func(ctx context.Context, fn func(ctx context.Context, id int) error) error"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: граничный — без ctx первым параметром
    Допустим тип "func(fn func(ctx context.Context) error) error"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: граничный — callback возвращает не error
    Допустим метод "func (r *JobRepository) NotInTx(ctx context.Context, fn func(ctx context.Context) (int, error)) error"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  # === Неприменимость ===

  Сценарий: неприменимость — анонимная tx-сигнатура вне service/usecase (в main)
    Допустим пакет "main" с функцией, принимающей параметр "tx func(ctx context.Context, fn func(ctx context.Context) error) error"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится
    # (Проверка 3 действует только в /domain/service и /domain/usecase.
    #  Проверка 1 ловит лишь именованные объявления типов, а не анонимные параметры.)

# --- Чек-лист при добавлении нового правила ---
#  [x] ID и описание занесены в реестр (RULES.md, GID-175)
#  [x] Выбран слой: go/analysis (анализатор gidintransaction в analyzers/intransaction)
#  [x] Заданы severity и сообщение ("GID-175: …")
#  [x] Покрыты кейсы: позитивный, негативный, граничный, неприменимость
#  [x] testdata с // want для analysistest
#  [ ] Правило включено в .golangci.yml
