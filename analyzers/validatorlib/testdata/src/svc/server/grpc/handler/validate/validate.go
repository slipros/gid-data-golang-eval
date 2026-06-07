// Негатив: validate-пакет использует validator.go — норма (grpc-запросы).
package validate

import (
	"context"

	validator "github.com/raoptimus/validator.go/v2"
)

func CreateSnapshot(ctx context.Context, in any) error {
	return validator.ValidateStruct(ctx, in)
}
