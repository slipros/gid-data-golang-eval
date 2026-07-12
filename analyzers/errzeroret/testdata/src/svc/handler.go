// Eval of GID-243: on error, non-error results must be nil/zero.
package svc

import (
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

// --- Non-applicability: a single-result return (no other result to check) ---

func single() error {
	return status.Error(codes.Internal, "x")
}
