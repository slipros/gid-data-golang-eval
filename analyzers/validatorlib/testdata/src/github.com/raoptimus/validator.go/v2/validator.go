// Stub validator.go/v2 для eval.
package validator

import "context"

type Rule interface{}

func ValidateStruct(ctx context.Context, v any, rules ...Rule) error { return nil }
