// Eval for GID-126: negative in the app layer — a bare Options and composition
// are the norm, but defaults must be Default* here as well.
package config

// --- Negative: a bare Options in the app layer (composition) — the norm ---

type Options struct {
	GRPC  GRPCOptions
	Kafka KafkaOptions
}

type GRPCOptions struct {
	Addr string
}

type KafkaOptions struct {
	Brokers []string
}

// --- Negative: defaults with the Default prefix — ok in the app layer too ---

var DefaultGRPCOptions = GRPCOptions{Addr: ":50051"}

// --- Positive: a default without the Default prefix is checked in the app layer too ---

var kafkaOpts = KafkaOptions{Brokers: []string{"localhost:9092"}} // want `GID-126: option defaults must be a Default<X>Options variable\. Fix: rename it`
