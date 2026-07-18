// Non-applicability (GID-138): a nested dal/repository below another layer
// (a server-side package at server/grpc/dal/repository/redis) is NOT the
// dal/repository layer — the layer root is anchored to the first segments
// after the module root ("server"). No grouping-subpackage diagnostic here.
package redis

type Cache struct{}
