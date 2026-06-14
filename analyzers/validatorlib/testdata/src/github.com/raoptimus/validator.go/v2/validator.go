// A stub of validator.go/v2 for the eval.
package validator

import "context"

type Rule interface{}

func ValidateStruct(ctx context.Context, v any, rules ...Rule) error { return nil }
