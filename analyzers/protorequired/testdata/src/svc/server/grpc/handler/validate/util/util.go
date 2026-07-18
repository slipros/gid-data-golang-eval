// Eval GID-232: boundary — a subpackage nested under a validate package
// (handler/validate/util) is not itself the validate package: "validate" is
// not the trailing (leaf) segment here. Without the EndsWith fix,
// pathseg.Contains would have falsely put this package in scope and flagged
// the NewRequired call on the enum field Executor below.
package util

import (
	"context"

	validator "github.com/raoptimus/validator.go/v2"

	"gen/commonpb"
)

// Boundary class: a validator whose package merely nests under validate/, not
// the validate leaf package itself — the rule does not apply, even though the
// RuleSet uses NewRequired on a proto3 enum field.
type CreateStage struct {
	rules validator.RuleSet
}

func NewCreateStage() *CreateStage {
	return &CreateStage{
		rules: validator.RuleSet{
			"Executor": {validator.NewRequired()},
		},
	}
}

func (v *CreateStage) Validate(ctx context.Context, req *commonpb.CreateStageRequest) error {
	return validator.ValidateStruct(ctx, req, v.rules)
}
