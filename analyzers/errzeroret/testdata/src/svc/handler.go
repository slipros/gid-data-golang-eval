// Eval of GID-243: on error, non-error results must be nil/zero.
package svc

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"svc/model"
	"svc/pb"
)

// --- Positive (a): a constructing error call, a populated (non-zero) other result ---

func badConstructing() (*pb.Resp, error) {
	return &pb.Resp{}, status.Error(codes.Internal, "x") // want `GID-243: on error, non-error results must be nil/zero`
}

// --- Positive (b): return inside `if err != nil`, a variable (non-zero) other result ---

func call() (int, error) { return 0, nil }

func badGuarded() (int, error) {
	res, err := call()
	if err != nil {
		return res, err // want `GID-243: on error, non-error results must be nil/zero`
	}
	return res, nil
}

// --- Negative: nil alongside a constructing error — ok ---

func goodNilConstructing() (*pb.Resp, error) {
	return nil, status.Error(codes.Internal, "x")
}

// --- Negative: a zero composite literal inside an if-guard — ok ---

func callResult() (model.Result, error) { return model.Result{}, nil }

func goodZeroGuarded() (model.Result, error) {
	res, err := callResult()
	if err != nil {
		return model.Result{}, err
	}
	return res, nil
}

// --- Negative: an unconditional final forward (interceptor pass-through) — ok ---

func handler(req int) (int, error) { return req, nil }

func forward(req int) (int, error) {
	resp, err := handler(req)
	return resp, err
}

// --- Negative: a nil error alongside a variable result — ok ---

func goodNilErr() (int, error) {
	res := 42
	return res, nil
}

// --- Negative: a zero-valued enum constant alongside a constructing error — ok.
// The proto *_UNSPECIFIED member is the enum's zero value (const 0), written as
// a selector rather than a literal — semantically "returned zero on error". ---

func goodProtoUnspecified(s string) (pb.Status, error) {
	return pb.Status_STATUS_UNSPECIFIED, errors.WithStack(errors.New("unhandled: " + s))
}

// --- Negative: a string-based enum's zero member ("") alongside an error — ok ---

func goodStringEnumUnspecified() (model.TranscribeJobSource, error) {
	return model.TranscribeJobSourceUnspecified, errors.WithStack(errors.New("x"))
}

// --- Negative: an int-based enum's zero member (0) alongside a guarded error — ok ---

func priorityCall() (model.Priority, error) { return model.PriorityUnspecified, nil }

func goodIntEnumUnspecifiedGuarded() (model.Priority, error) {
	p, err := priorityCall()
	if err != nil {
		return model.PriorityUnspecified, err
	}
	return p, nil
}

// --- Positive: a NON-zero enum constant alongside an error is still flagged.
// A non-zero constant is not the zero value — the zero-const relaxation must
// not leak to it. ---

func badNonZeroEnum() (pb.Status, error) {
	return pb.Status_STATUS_ACTIVE, status.Error(codes.Internal, "x") // want `GID-243: on error, non-error results must be nil/zero`
}

func badNonZeroStringEnum() (model.TranscribeJobSource, error) {
	return model.TranscribeJobSourceUpload, errors.WithStack(errors.New("x")) // want `GID-243: on error, non-error results must be nil/zero`
}

// --- Non-applicability: a single-result return (no other result to check) ---

func single() error {
	return status.Error(codes.Internal, "x")
}
