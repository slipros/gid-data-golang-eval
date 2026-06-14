// Eval GID-232: NewRequired on proto3 enum fields in the validate layer.
package validate

import (
	"context"

	validator "github.com/raoptimus/validator.go/v2"

	"gen/commonpb"
)

// --- Positive class: violations ---

// CreateStage — rules built in the constructor, request type resolved from
// the Validate method of the constructed type.
type CreateStage struct {
	rules validator.RuleSet
}

func NewCreateStage() *CreateStage {
	return &CreateStage{
		rules: validator.RuleSet{
			// Negative (boundary): NewRequired on a string field — ok.
			"Name": {validator.NewRequired()},
			// Positive: NewRequired on a proto3 enum field.
			"Executor": {validator.NewRequired()}, // want `GID-232: NewRequired on proto3 enum field "Executor" treats \*_UNSPECIFIED=0 as empty\. Fix: validator\.NewInRange\(pb\.Status_ACTIVE, pb\.Status_CLOSED\)`
			// Positive (boundary): chained NewRequired().When(...) on an enum.
			"Input": {
				validator.NewRequired(), // Input is a *message — ok
				validator.NewNested(validator.RuleSet{
					// Positive: nested rule set, enum field of the nested message.
					"Source": {validator.NewRequired()}, // want `GID-232: NewRequired on proto3 enum field "Source" treats \*_UNSPECIFIED=0 as empty\. Fix: validator\.NewInRange\(pb\.Status_ACTIVE, pb\.Status_CLOSED\)`
				}).SkipOnEmpty(),
			},
			// Positive: NewEach(NewNested(...)) — enum field of a repeated message.
			"Stages": {
				validator.NewEach(
					validator.NewNested(validator.RuleSet{
						"Name":     {validator.NewRequired()},
						"Executor": {validator.NewRequired()}, // want `GID-232: NewRequired on proto3 enum field "Executor" treats \*_UNSPECIFIED=0 as empty\. Fix: validator\.NewInRange\(pb\.Status_ACTIVE, pb\.Status_CLOSED\)`
					}),
				).SkipOnEmpty(),
			},
			// Negative (boundary): pointer enum (proto3 optional) — nil is
			// genuinely empty, NewRequired works, not flagged.
			"OptionalExecutor": {validator.NewRequired()},
			// Negative (boundary): named int32 without proto enum methods.
			"Priority": {validator.NewRequired()},
			// Non-applicability: key is not a field of the request — skipped.
			"Unknown": {validator.NewRequired()},
		},
	}
}

func (v *CreateStage) Validate(ctx context.Context, req *commonpb.CreateStageRequest) error {
	return validator.ValidateStruct(ctx, req, v.rules)
}

// UpdateStage — RuleSet literal directly inside the Validate method.
type UpdateStage struct{}

func (v *UpdateStage) Validate(ctx context.Context, req *commonpb.UpdateStageRequest) error {
	whenAlways := func(ctx context.Context, value any) bool { return true }
	rules := validator.RuleSet{
		// Positive (boundary): chained NewRequired().When(...) on an enum.
		"Executor": {validator.NewRequired().When(whenAlways)}, // want `GID-232: NewRequired on proto3 enum field "Executor" treats \*_UNSPECIFIED=0 as empty\. Fix: validator\.NewInRange\(pb\.Status_ACTIVE, pb\.Status_CLOSED\)`
		// Negative: NewInRange on a proto3 enum field — the mandated fix.
		"Status": {validator.NewInRange([]any{
			commonpb.StageExecutor_STAGE_EXECUTOR_AGENT,
			commonpb.StageExecutor_STAGE_EXECUTOR_REPLICATOR,
		})},
	}
	return validator.ValidateStruct(ctx, req, rules)
}

// --- Non-applicability: RuleSet without a resolvable validated struct ---

// sharedRules has no Validate(ctx, req) context — the validated struct
// cannot be resolved confidently, so the rule set is skipped (FP-safe).
func sharedRules() validator.RuleSet {
	return validator.RuleSet{
		"Executor": {validator.NewRequired()},
	}
}

var _ = sharedRules
