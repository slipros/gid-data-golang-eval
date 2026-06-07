# language: ru

Функция: GID-126 — Options-паттерн: имя типа и дефолты (optsnaming)
  Как разработчик
  Я хочу, чтобы тип настроек именовался с префиксом сущности (JobOptions),
  а дефолты жили в переменной Default<X>Options
  Чтобы Options-паттерн был единообразен; в app-слое голый Options — норма (композиция)

  # Анализатор gidoptsnaming, LoadMode по умолчанию (TypesInfo).
  # Слой app определяется по сегменту пути "app" через internal/pathseg.
  # Сгенерированный код (ast.IsGenerated) пропускается.
  #
  # Проверки:
  #   1. struct-тип с именем РОВНО Options вне app-слоя.
  #      В app-слое голый Options — норма (агрегирует GRPCOptions/KafkaOptions), FINDINGS.md §2.3.
  #      Не-struct типы (alias на Options, interface) не задеваются.
  #   2. package-level var типа <X>Options (включая указатель), имя не начинается с Default.
  #      Локальные переменные не задеваются. Var в app-слое тоже проверяется.
  #
  # Соседнее правило GID-152 (gidoptsstyle) проверяет указатель/embedding opts —
  # здесь не дублируется.

  # === Класс 1: позитив (нарушение ловится) ===

  Сценарий: позитивный — struct Options вне app-слоя
    Допустим пакет в "/domain/service" с типом "type Options struct { Retries int }"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-126: тип настроек — с префиксом сущности: JobOptions, не голый Options" на типе "Options"

  Сценарий: позитивный — package-level var типа JobOptions без префикса Default
    Допустим пакет в "/domain/service" с "var Opts = JobOptions{}"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-126: дефолты Options — переменная Default<X>Options" на переменной "Opts"

  Сценарий: позитивный — package-level var с явным типом JobOptions без Default
    Допустим пакет в "/domain/service" с "var defaults JobOptions"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-126: дефолты Options — переменная Default<X>Options" на переменной "defaults"

  Сценарий: позитивный — дефолт без префикса Default проверяется и в app-слое
    Допустим пакет в "/internal/app/config" с "var kafkaOpts = KafkaOptions{}"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-126: дефолты Options — переменная Default<X>Options" на переменной "kafkaOpts"

  # === Класс 2: негатив (чистый код проходит) ===

  Сценарий: негативный — тип с префиксом сущности JobOptions
    Допустим пакет в "/domain/service" с типом "type JobOptions struct { Retries int }"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: негативный — дефолты в переменной Default<X>Options
    Допустим пакет в "/domain/service" с "var DefaultJobOptions = JobOptions{}"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: негативный — голый Options в app-слое (композиция GRPCOptions/KafkaOptions)
    Допустим пакет в "/internal/app/config" с типом "type Options struct { GRPC GRPCOptions; Kafka KafkaOptions }"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  # === Класс 3: граничный (похожее, но не флагуем) ===

  Сценарий: граничный — локальная переменная opts не матчится
    Допустим функцию с локальной "opts := JobOptions{}"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: граничный — var-указатель с префиксом Default — ок
    Допустим пакет в "/domain/service" с "var DefaultGRPCOptions *GRPCOptions"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: граничный — функция с параметром opts — не зона этого правила
    Допустим функцию "func New(ctx context.Context, opts *JobOptions) int"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится
    # (Параметры/поля opts проверяет GID-152, не GID-126.)

  Сценарий: граничный — alias на Options и interface не задеваются
    Допустим пакет в "/domain/model" с "type Options = entOptions" и "type OptionsProvider interface { ... }"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится
    # (Проверка 1 действует только на struct-типы с именем ровно Options.)

  # === Класс 4: неприменимость ===

  Сценарий: неприменимость — пакет без Options-типов
    Допустим пакет в "/domain/model" с типом "type Job struct { ID int }" и "var DefaultJob = Job{}"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

# --- Чек-лист при добавлении нового правила ---
#  [x] ID и описание занесены в реестр (RULES.md, GID-126)
#  [x] Выбран слой: go/analysis (анализатор gidoptsnaming в analyzers/optsnaming)
#  [x] Заданы severity и сообщение ("GID-126: …")
#  [x] Покрыты кейсы: позитивный, негативный, граничный, неприменимость
#  [x] testdata с // want для analysistest
#  [ ] Правило включено в .golangci.yml
