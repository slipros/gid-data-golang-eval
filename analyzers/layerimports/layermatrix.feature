# language: ru

Функция: GID-224…229 — матрица изоляции слоёв
  Как архитектор сервиса
  Я хочу, чтобы каждый слой импортировал только то, что ему положено
  Чтобы зависимости текли в одну сторону, а wiring жил в composition root

  # GID-224: транспорт видит только domain/model (и validate)

  Сценарий: server импортирует domain/service — нарушение
    Допустим пакет "svc/server/http/handler" импортирует "svc/domain/service"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-224" на импорте "svc/domain/service"

  Сценарий: server импортирует dal/repository — нарушение
    Допустим пакет "svc/server/http/handler" импортирует "svc/dal/repository"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-224" на импорте "svc/dal/repository"

  Сценарий: schedule импортирует domain/service — нарушение
    Допустим пакет "svc/schedule/sync" импортирует "svc/domain/service"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-224" на импорте "svc/domain/service"

  Сценарий: validate импортирует dal/entity — нарушение
    Допустим пакет "svc/validate" импортирует "svc/dal/entity"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-224" на импорте "svc/dal/entity"

  Сценарий: event-consumer импортирует dal/entity и domain/service — нарушения
    Допустим пакет "svc/event/consumer" импортирует "svc/dal/entity" и "svc/domain/service"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-224" на обоих импортах

  Сценарий: транспорт импортирует domain/model и validate — ок
    Допустим пакет "svc/server/http/handler" импортирует "svc/domain/model" и "svc/validate"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  # GID-225: composition root и транспорт — листья

  Сценарий: domain импортирует app — нарушение
    Допустим пакет "svc/domain/notifier" импортирует "svc/app"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-225" на импорте "svc/app"

  Сценарий: domain импортирует server — нарушение
    Допустим пакет "svc/domain/notifier" импортирует "svc/server/middleware"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-225" на импорте "svc/server/middleware"

  Сценарий: app импортирует все слои — ок (composition root)
    Допустим пакет "svc/app" импортирует repository, service, producer, client и metric
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  # GID-226: metric самостоятелен

  Сценарий: metric импортирует domain/model — нарушение
    Допустим пакет "svc/metric" импортирует "svc/domain/model"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-226" на импорте "svc/domain/model"

  Сценарий: domain/service импортирует metric — нарушение
    Допустим пакет "svc/domain/service" импортирует "svc/metric"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-226" на импорте "svc/metric"

  Сценарий: domain импортирует пакет с сегментом "metrics" — граница, ок
    Допустим пакет "svc/domain/boundary" импортирует "svc/metrics/registry"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  # GID-227: model — чистый словарь

  Сценарий: model импортирует транспорт — нарушение
    Допустим пакет "svc/domain/model" импортирует "svc/server/middleware"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-227" на импорте "svc/server/middleware"

  Сценарий: usecase импортирует подпакет model — ок (model-слой)
    Допустим пакет "svc/domain/usecase" импортирует "svc/domain/model/filter"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  # GID-228: клиент — через интерфейс в domain/model

  Сценарий: domain/service импортирует client — нарушение
    Допустим пакет "svc/domain/service" импортирует "svc/client/billing"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-228" на импорте "svc/client/billing"

  Сценарий: dal/repository импортирует client — нарушение
    Допустим пакет "svc/dal/repository" импортирует "svc/client/billing"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-228" на импорте "svc/client/billing"

  # GID-229: клиент изолирован

  Сценарий: client импортирует domain/model — нарушение
    Допустим пакет "svc/client/billing" импортирует "svc/domain/model"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-229" на импорте "svc/domain/model"

  Сценарий: client импортирует сторонний пакет — правило не применяется
    Допустим пакет "svc/client/billing" импортирует "strconv"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  # Сторонние библиотеки и настройки

  Сценарий: слой импортирует стороннюю библиотеку с сегментом "client" — ок
    Допустим импорт принадлежит другому модулю (префикс модуля не совпадает)
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: правило отключено через settings.disable — диагностики нет
    Допустим в .golangci.yml задано settings.disable: [GID-224]
    И пакет "custom/server/handler" импортирует "custom/domain/service"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: своё правило через settings.rules — диагностика с его ID
    Допустим в .golangci.yml задано правило SVC-1: scope "domain/service", banned [legacy]
    И пакет "custom/domain/service" импортирует "custom/legacy/store"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "SVC-1" на импорте "custom/legacy/store"

# --- Чек-лист при добавлении нового правила ---
#  [x] ID и описание занесены в реестр RULES.md
#  [x] Выбран слой: go/analysis (нужны сегменты import-пути)
#  [x] Заданы severity и сообщение
#  [x] Покрыты кейсы: позитивный, негативный, граничный, неприменимость
#  [x] testdata с // want для analysistest
#  [x] Правило включено в .golangci.yml
