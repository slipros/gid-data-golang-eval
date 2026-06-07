# language: ru

Функция: GID-188 — запрет кастомных context-типов (no-custom-context)
  Как разработчик
  Я хочу, чтобы в позиции ctx и в embedding интерфейсов использовался только context.Context
  Чтобы не плодить кастомные контексты и передавать данные через context.WithValue

  # Правило Google: "custom contexts — no exceptions".
  # Анализатор gidcustomctx, LoadMode TypesInfo. Детект (pass.TypesInfo):
  #   1. объявление именованного типа (struct/interface) в проверяемом пакете,
  #      method set которого покрывает context.Context (Deadline/Done/Err/Value)
  #      — через types.Implements относительно интерфейса stdlib context.Context;
  #   2. interface-тип, ВСТРАИВАЮЩИЙ context.Context (embedded в декларации);
  #   3. параметр функции/функционального литерала с именем ctx, чей тип —
  #      именованный не-stdlib тип (не context.Context).
  # Интерфейс context.Context берётся из импортов пакета (прямых/транзитивных);
  # если context нигде не импортируется, кейсы 1 и 2 неприменимы.
  # Сгенерированный код (ast.IsGenerated) пропускается.
  # Точечное отключение — стандартный //nolint:gidcustomctx.

  # === Класс 1: позитивные (нарушения) ===

  Сценарий: позитивный — interface встраивает context.Context
    Допустим объявление "type MyContext interface { context.Context; Extra() string }"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-188: кастомный context-тип MyContext запрещён — передавайте context.Context и кладите данные через context.WithValue (хелперы в /domain/model — GID-165/166)"

  Сценарий: позитивный — struct с полным набором методов context.Context
    Допустим тип "CtxStruct" с методами "Deadline/Done/Err/Value" с сигнатурами context.Context
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-188: кастомный context-тип CtxStruct запрещён ..."

  Сценарий: позитивный — параметр ctx кастомного типа
    Допустим функцию "func f(ctx MyCtx)" где MyCtx — именованный не-stdlib тип
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-188: параметр ctx имеет тип <MyCtx> — используйте context.Context"

  # === Класс 2: негативные (чистый код) ===

  Сценарий: негативный — параметр ctx это context.Context
    Допустим функцию "func f(ctx context.Context)"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: негативный — struct с методом Done, но не полным набором
    Допустим тип "PartialCtx" с единственным методом "Done() <-chan struct{}"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится
    # Method set не покрывает context.Context — types.Implements false.

  # === Класс 3: граничные ===

  Сценарий: граничный — interface { context.Context } матчится один раз
    Допустим объявление "type OnlyEmbed interface { context.Context }"
    Когда анализатор проверяет файл
    Тогда выводится ровно одна диагностика "GID-188: кастомный context-тип OnlyEmbed запрещён ..."
    # Кейс embedding (2) срабатывает и предотвращает повторный матч по types.Implements (1).

  Сценарий: граничный — методы Deadline/Done/Err/Value с другими сигнатурами
    Допустим тип "FakeCtx" с методами "Deadline() string, Done() bool, Err() string, Value(int) int"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится
    # Имена совпадают, но сигнатуры не равны context.Context — не матчится.

  Сценарий: граничный — параметр ctx stdlib-типа рядом с не-ctx параметром
    Допустим функцию "func f(ctx context.Context, other FakeCtx)"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится
    # Имя ctx закреплено за context.Context; параметр other не проверяется по имени.

  # === Класс 4: неприменимость ===

  Сценарий: неприменимость — пакет без context
    Допустим пакет, который не импортирует context и не имеет context-подобных типов
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

# --- Чек-лист при добавлении нового правила ---
#  [x] ID и описание занесены в реестр (RULES.md, GID-188)
#  [x] Выбран слой: go/analysis (анализатор gidcustomctx в analyzers/customctx)
#  [x] Заданы severity и сообщение ("GID-188: …")
#  [x] Покрыты кейсы: позитивный, негативный, граничный, неприменимость
#  [x] testdata с // want для analysistest
#  [ ] Правило включено в .golangci.yml
