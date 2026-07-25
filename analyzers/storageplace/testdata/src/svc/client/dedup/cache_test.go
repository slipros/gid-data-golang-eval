// Boundary: a _test.go file may talk to the driver from any layer
// (an integration test, a miniredis fixture).
package dedup

import redis "github.com/redis/go-redis/v9"

var testClient *redis.Client
