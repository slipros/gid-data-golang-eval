// Eval для GID-152 (opts style).
package optsstyle

type HelloOptions struct {
	Retries int
}

var DefaultHelloOptions = HelloOptions{Retries: 3}

// --- Позитив: opts по значению в параметре ---

func NewBad(opts HelloOptions) int { // want `GID-152: opts передаётся указателем — используйте \*HelloOptions`
	return opts.Retries
}

// --- Позитив: именованное поле вместо встраивания ---

type BadHello struct {
	opts HelloOptions // want `GID-152: opts встраивается в тело сущности \(embedded HelloOptions\), а не хранится именованным полем`
}

// Граничный кейс: именованное поле-указатель — тоже нарушение.
type BadPtrHello struct {
	opts *HelloOptions // want `GID-152: opts встраивается в тело сущности \(embedded HelloOptions\), а не хранится именованным полем`
}

// --- Негатив: указатель в параметре, embedded в структуре ---

func NewGood(opts *HelloOptions) *GoodHello {
	return &GoodHello{HelloOptions: *opts}
}

type GoodHello struct {
	HelloOptions
}

// --- Неприменимость: тип без постфикса Options ---

type Config struct {
	Retries int
}

func WithConfig(cfg Config) int { return cfg.Retries }

type Holder struct {
	cfg Config
}
