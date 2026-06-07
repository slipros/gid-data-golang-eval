// Eval для GID-126: негатив в app-слое — голый Options и композиция норма,
// но дефолты и здесь должны быть Default*.
package config

// --- Негатив: голый Options в app-слое (композиция) — норма ---

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

// --- Негатив: дефолты с префиксом Default — ок и в app-слое ---

var DefaultGRPCOptions = GRPCOptions{Addr: ":50051"}

// --- Позитив: дефолт без префикса Default проверяется и в app-слое ---

var kafkaOpts = KafkaOptions{Brokers: []string{"localhost:9092"}} // want `GID-126: дефолты Options — переменная Default<X>Options`
