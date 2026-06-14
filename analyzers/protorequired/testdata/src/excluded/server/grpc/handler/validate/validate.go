// Non-applicability: enum fields listed in settings.exclude are not flagged.
package validate

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
			// "CreateStageRequest.Executor" is in settings.exclude — ok.
			"Executor": {validator.NewRequired()},
		},
	}
}

func (v *CreateStage) Validate(ctx context.Context, req *commonpb.CreateStageRequest) error {
	return validator.ValidateStruct(ctx, req, v.rules)
}
