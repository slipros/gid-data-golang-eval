// Package commonpb stubs generated protobuf code for the eval: a proto3
// enum is a named int32 with String() and EnumDescriptor() methods.
package commonpb

// StageExecutor — proto3 enum: named int32 + String + EnumDescriptor.
type StageExecutor int32

const (
	StageExecutor_STAGE_EXECUTOR_UNSPECIFIED StageExecutor = 0
	StageExecutor_STAGE_EXECUTOR_AGENT       StageExecutor = 1
	StageExecutor_STAGE_EXECUTOR_REPLICATOR  StageExecutor = 2
)

func (x StageExecutor) String() string { return "STAGE_EXECUTOR_UNSPECIFIED" }

func (StageExecutor) EnumDescriptor() ([]byte, []int) { return nil, nil }

// Priority — named int32 that is NOT a proto enum (no Stringer/descriptor).
type Priority int32

// StageInput — nested proto message.
type StageInput struct {
	Source StageExecutor
}

// Stage — repeated proto message element.
type Stage struct {
	Name     string
	Executor StageExecutor
}

// CreateStageRequest — proto message validated by the fixtures.
type CreateStageRequest struct {
	Name             string
	Executor         StageExecutor
	OptionalExecutor *StageExecutor
	Priority         Priority
	Input            *StageInput
	Stages           []*Stage
}

// UpdateStageRequest — proto message for the in-method RuleSet fixture.
type UpdateStageRequest struct {
	Executor StageExecutor
	Status   StageExecutor
}
