// Non-applicability: not a validate-layer package — the same code is not
// flagged outside packages with a "validate" path segment.
package service

import (
	"context"

	validator "github.com/raoptimus/validator.go/v2"

	"gen/commonpb"
)

type CreateStage struct {
	rules validator.RuleSet
}

func NewCreateStage() *CreateStage {
	return &CreateStage{
		rules: validator.RuleSet{
			"Executor": {validator.NewRequired()}, // not the validate layer — ok
		},
	}
}

func (v *CreateStage) Validate(ctx context.Context, req *commonpb.CreateStageRequest) error {
	return validator.ValidateStruct(ctx, req, v.rules)
}
