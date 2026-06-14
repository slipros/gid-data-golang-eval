// Stub of validator.go/v2 for the eval.
package validator

import "context"

type Rule interface{}

// RuleSet — rules per struct field.
type RuleSet map[string][]Rule

// WhenFunc — condition for conditional rules.
type WhenFunc func(ctx context.Context, value any) bool

type RequiredRule struct{}

func NewRequired() RequiredRule { return RequiredRule{} }

func (r RequiredRule) When(fn WhenFunc) RequiredRule { return r }

func (r RequiredRule) SkipOnEmpty() RequiredRule { return r }

type InRangeRule struct{}

func NewInRange(values any) InRangeRule { return InRangeRule{} }

func (r InRangeRule) SkipOnEmpty() InRangeRule { return r }

type NestedRule struct{}

func NewNested(rules RuleSet) NestedRule { return NestedRule{} }

func (r NestedRule) SkipOnEmpty() NestedRule { return r }

type EachRule struct{}

func NewEach(rules ...Rule) EachRule { return EachRule{} }

func (r EachRule) SkipOnEmpty() EachRule { return r }

func ValidateStruct(ctx context.Context, v any, rules RuleSet) error { return nil }
