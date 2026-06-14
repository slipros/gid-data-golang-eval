// Negative: the validate package uses validator.go — the norm (grpc requests).
package validate

import (
	"context"

	validator "github.com/raoptimus/validator.go/v2"
)

func CreateSnapshot(ctx context.Context, in any) error {
	return validator.ValidateStruct(ctx, in)
}
