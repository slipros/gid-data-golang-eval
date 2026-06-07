# language: ru

Функция: GID-114 — методы repo/service именуются от сущности (entitymethod)
  Как разработчик
  Я хочу, чтобы экспортируемые методы структур в /dal/repository и /domain/service
  именовались от сущности: без префикса List, без суффикса ByID, с именем сущности в имени
  Чтобы API слоя читался единообразно (Jobs вместо ListJobs, Job(ctx, id) вместо JobByID)

  # Один анализатор entitymethod → линтер gidentitymethod, LoadModeTypesInfo.
  # Scope — корневые пакеты слоя по сегментам пути (pathseg.EndsWith):
  #   /dal/repository и /domain/service. Подпакеты convert/build вне scope.
  # Проверяются только ЭКСПОРТИРУЕМЫЕ методы структур (fn.Recv != nil).
  # Конструкторы New* — функции без ресивера, сюда не попадают.
  # Сгенерированный код (ast.IsGenerated) пропускается.
  #
  # Три проверки (в порядке приоритета, репортится первая сработавшая):
  #   1. префикс List по границе CamelCase запрещён → «без префикса List …»;
  #   2. точный суффикс ByID запрещён → «без суффикса ByID …»
  #      (ByStageID и прочие By<Field>ID разрешены — уточнение выборки);
  #   3. имя метода обязано содержать имя типа-ресивера как CamelCase-подстроку.
  #      Применяется только к осмысленному имени сущности (len > 2);
  #      однобуквенные/служебные ресиверы (T, S, ID) не проверяются.
  #
  # FP-зона: методы-глаголы без имени сущности (Close, Ping, Flush) в репозитории
  # легитимны редко, но бывают — они попадут под проверку 3 и гасятся
  # //nolint:gidentitymethod или settings.exclude ("Метод" | "Тип.Метод").

  # --- Класс 1: позитивный (нарушение ловится) ---

  Сценарий: позитивный — префикс List
    Допустим метод "func (j *Job) ListJobs(ctx context.Context) ([]Snapshot, error)" в /dal/repository
    Когда анализатор gidentitymethod проверяет файл
    Тогда выводится диагностика "GID-114: без префикса List — множественное число: Jobs вместо ListJobs"

  Сценарий: позитивный — суффикс ByID
    Допустим метод "func (j *Job) JobByID(ctx context.Context, id string) (Snapshot, error)" в /dal/repository
    Когда анализатор gidentitymethod проверяет файл
    Тогда выводится диагностика "GID-114: без суффикса ByID — Job(ctx, id) вместо JobByID"

  Сценарий: позитивный — имя метода не содержит сущность
    Допустим метод "func (j *Job) Fetch(ctx context.Context) (Snapshot, error)" в /dal/repository
    Когда анализатор gidentitymethod проверяет файл
    Тогда выводится диагностика "GID-114: имя метода \"Fetch\" должно содержать имя сущности \"Job\""

  Сценарий: позитивный — метод-глагол без сущности (FP-зона)
    Допустим метод "func (j *Job) Close() error" в /dal/repository
    Когда анализатор gidentitymethod проверяет файл
    Тогда выводится диагностика "GID-114: имя метода \"Close\" должно содержать имя сущности \"Job\""
    # Это и есть FP-зона: легитимный Close/Ping/Flush ловится — выключается exclude/nolint.

  # --- Класс 2: негативный (чистый код проходит) ---

  Сценарий: негативный — имена от сущности
    Допустим методы "Job", "Jobs", "CreateJob", "DeleteJob" на типе Job в /dal/repository
    Когда анализатор gidentitymethod проверяет файл
    Тогда диагностика не выводится

  # --- Класс 3: граничный (похоже на нарушение, но допустимо) ---

  Сценарий: граничный — суффикс ByStageID разрешён
    Допустим метод "func (j *Job) JobsByStageID(ctx context.Context, stageID string) ([]Snapshot, error)" в /dal/repository
    Когда анализатор gidentitymethod проверяет файл
    Тогда диагностика не выводится
    # ByStageID — уточнение выборки (By<Field>ID), не точный суффикс ByID; имя содержит Job.

  Сценарий: граничный — Listen не считается префиксом List
    Допустим метод "func (j *Job) ListenJobEvents(ctx context.Context) error" в /dal/repository
    Когда анализатор gidentitymethod проверяет файл
    Тогда диагностика не выводится
    # Граница CamelCase: после "List" идёт строчная "e" — это не слово List.

  Сценарий: граничный — неэкспортируемый метод не матчится
    Допустим метод "func (j *Job) listJobsInternal(ctx context.Context) ([]Snapshot, error)" в /dal/repository
    Когда анализатор gidentitymethod проверяет файл
    Тогда диагностика не выводится

  Сценарий: граничный — однобуквенная сущность: проверка 3 не применяется
    Допустим тип "S" и метод "func (x *S) Touch(ctx context.Context) error" в /domain/service
    Когда анализатор gidentitymethod проверяет файл
    Тогда диагностика не выводится
    # len(recv) <= 2 — имя сущности служебное, требование «содержать сущность» не проверяется.
    # Префикс List/суффикс ByID при этом всё равно ловятся (от длины не зависят).

  Сценарий: граничный — settings.exclude гасит метод-глагол
    Допустим settings.exclude = ["Job.Close", "Ping"] и методы Close, Ping, Flush на типе Job
    Когда анализатор gidentitymethod проверяет файл
    Тогда диагностика выводится только на Flush; Close и Ping погашены

  # --- Класс 4: неприменимость ---

  Сценарий: неприменимость — подпакет convert вне scope
    Допустим метод "ListSnapshots" на типе Mapper в /dal/repository/convert
    Когда анализатор gidentitymethod проверяет файл
    Тогда диагностика не выводится
    # Scope — корень слоя (EndsWith), подпакеты convert/build не задеваются.

  Сценарий: неприменимость — /domain/usecase вне scope
    Допустим методы "ListJobs" и "Fetch" на типе Job в /domain/usecase
    Когда анализатор gidentitymethod проверяет файл
    Тогда диагностика не выводится
    # Scope правила — только repository и service; usecase не задевается.

# --- Чек-лист при добавлении нового правила ---
#  [x] ID и описание занесены в реестр (RULES.md, GID-114)
#  [x] Выбран слой: go/analysis (пакет entitymethod: gidentitymethod), LoadModeTypesInfo
#  [x] Задано сообщение ("GID-114: …")
#  [x] Покрыты кейсы: позитивный, негативный, граничный, неприменимость
#  [x] testdata с // want для analysistest
#  [ ] Правило включено в .golangci.yml
