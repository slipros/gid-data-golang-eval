// Eval для GID-126: позитивные и граничные кейсы вне app-слоя (/domain/service).
package service

import "context"

// --- Позитив: struct-тип с именем ровно Options вне app-слоя ---

type Options struct { // want `GID-126: тип настроек — с префиксом сущности: JobOptions, не голый Options`
	Retries int
}

// Сущностный тип Options — используется дефолтами и параметрами ниже.
type JobOptions struct {
	Retries int
}

// --- Позитив: package-level var типа <X>Options без префикса Default ---

var Opts = JobOptions{Retries: 3} // want `GID-126: дефолты Options — переменная Default<X>Options`

// --- Позитив: package-level var-объявление (тип явно указан) без Default ---

var defaults JobOptions // want `GID-126: дефолты Options — переменная Default<X>Options`

// --- Негатив: дефолты в переменной Default<X>Options ---

var DefaultJobOptions = JobOptions{Retries: 5}

// --- Граничный: локальная переменная opts — не матчится ---

func use() int {
	opts := JobOptions{Retries: 1}
	return opts.Retries
}

// --- Граничный: функция с параметром opts — не зона этого правила ---

func New(ctx context.Context, opts *JobOptions) int {
	_ = ctx
	return opts.Retries
}

// --- Граничный: var-указатель с префиксом Default — ок ---

var DefaultGRPCOptions *GRPCOptions

type GRPCOptions struct {
	Addr string
}
