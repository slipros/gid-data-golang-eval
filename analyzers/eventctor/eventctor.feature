# language: ru
# GID-216 — event-ctor-deps (линтер gideventctor). Источник: event.md.

Функция: GID-216 — зависимости конструкторов event-слоя
  Как разработчик event-слоя
  Я хочу, чтобы конструкторы consumer'ов принимали logrus-logger,
  а конструкторы producer'ов — нет
  Чтобы consumer собирал Entry с полями broker/consumer, а producer
  пробрасывал ошибки вызывающему коду

  Сценарий: consumer-конструктор без logger — нарушение (позитив)
    Допустим пакет с сегментами event и consumer
    И конструктор "func NewOrderConsumer(svc Service) *OrderConsumer"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-216" о том, что consumer принимает *logrus.Logger

  Сценарий: producer-конструктор с *logrus.Logger — нарушение (позитив)
    Допустим пакет с сегментами event и producer
    И конструктор "func NewOrderProducer(log *logrus.Logger) *OrderProducer"
    Когда анализатор проверяет файл
    Тогда выводится диагностика "GID-216" о том, что producer не принимает logger

  Сценарий: consumer-конструктор с *logrus.Logger — ок (негатив)
    Допустим пакет с сегментами event и consumer
    И конструктор "func NewPaymentConsumer(log *logrus.Logger) *PaymentConsumer"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: consumer-конструктор с *logrus.Entry — ок (негатив)
    Допустим пакет с сегментами event и consumer
    И конструктор "func NewRefundConsumer(log *logrus.Entry) *RefundConsumer"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: producer-конструктор без logger — ок (негатив)
    Допустим пакет с сегментами event и producer
    И конструктор "func NewPaymentProducer(svc Service) *PaymentProducer"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: schema-функция возвращает тип чужого пакета — не конструктор (граничный)
    Допустим пакет с сегментами event и consumer
    И функция "func NewOrderCreatedSchema() *registry.Schema" возвращает тип чужого пакета
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: неэкспортируемый хелпер — не конструктор (граничный)
    Допустим пакет с сегментами event и consumer
    И функция "func newHelper() *helper"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: валидатор в event/kafka/consumer/validate — не consumer (граничный)
    Допустим пакет с сегментами event, consumer и validate
    И конструктор "func NewOrderValidator() *OrderValidator"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: конструктор вне event-слоя — правило не применяется (неприменимость)
    Допустим пакет в слое domain/service
    И конструктор "func NewService() *Service"
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

  Сценарий: конструктор из settings.exclude — пропускается (неприменимость)
    Допустим пакет с сегментами event и consumer
    И конструктор "func NewLegacyConsumer() *LegacyConsumer" числится в settings.exclude
    Когда анализатор проверяет файл
    Тогда диагностика не выводится

# --- Чек-лист при добавлении нового правила ---
#  [x] ID и описание занесены в реестр (RULES.md)
#  [x] Выбран слой: go/analysis (сложное — нужен types)
#  [x] Заданы severity и сообщение
#  [x] Покрыты кейсы: позитивный, негативный, граничный, неприменимость
#  [x] testdata с // want для analysistest
#  [x] Правило включено в .golangci.yml
