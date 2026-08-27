// Eval for GID-271: the consumer of the generated genports.go — the
// generated port file stays silent despite exactly one consumer.
package grpc

// CodecServer is the only consumer of the generated Codec interface.
type CodecServer struct {
	codec Codec
}
