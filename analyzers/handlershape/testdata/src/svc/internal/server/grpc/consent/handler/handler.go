// Eval GID-230: handler shape in a handler package under /server.
package handler

import (
	"context"
	"io"
)

// docReq/docResp — proto request/response stand-ins.
type docReq struct{ ID string }
type docResp struct{ OK bool }

// --- Negative class: canonical handler ---

type DocumentsValidator interface {
	Validate(ctx context.Context, req *docReq) error
}

type DocumentsService interface {
	Documents(ctx context.Context, id string) (*docResp, error)
}

// Canonical handler: Handle(ctx, req) (resp, error), deps named after the
// handler; a non-interface field (timeout) is not checked.
type Documents struct {
	validator DocumentsValidator
	service   DocumentsService
	timeout   int
}

func (h *Documents) Handle(ctx context.Context, req *docReq) (*docResp, error) {
	if err := h.validator.Validate(ctx, req); err != nil {
		return nil, err
	}
	return h.service.Documents(ctx, req.ID)
}

// Settings type — not a handler, not flagged.
type ListOptions struct{ Limit int }

// Unexported struct — not flagged.
type helper struct{} //nolint:unused

// --- Positive class: violations ---

// No Handle method at all.
type Purge struct{} // want `GID-230: handler "Purge" must have a Handle method with context\.Context as the first param and error as the last result\. Fix: func \(h \*Purge\) Handle\(ctx context\.Context, req \*rpc\.Request\) \(\*rpc\.Response, error\)`

// Handle without ctx as the first parameter.
type Update struct{} // want `GID-230: handler "Update" must have a Handle method with context\.Context as the first param and error as the last result\. Fix: func \(h \*Update\) Handle\(ctx context\.Context, req \*rpc\.Request\) \(\*rpc\.Response, error\)`

func (h *Update) Handle(req *docReq) error { return nil }

// Handle whose last result is not error.
type Remove struct{} // want `GID-230: handler "Remove" must have a Handle method with context\.Context as the first param and error as the last result\. Fix: func \(h \*Remove\) Handle\(ctx context\.Context, req \*rpc\.Request\) \(\*rpc\.Response, error\)`

func (h *Remove) Handle(ctx context.Context, req *docReq) (*docResp, bool) { return nil, false }

// Interface dependency named after another handler.
type ExportService interface {
	Export(ctx context.Context, id string) (*docResp, error)
}

type Export struct {
	validator DocumentsValidator // want `GID-230: handler "Export" interface dependency "DocumentsValidator" must be named "ExportValidator" or "ExportService"\. Fix: type ExportValidator interface\{ Validate\(ctx context\.Context, req \*rpc\.Request\) error \}`
	service   ExportService
}

func (h *Export) Handle(ctx context.Context, req *docReq) (*docResp, error) {
	if err := h.validator.Validate(ctx, req); err != nil {
		return nil, err
	}
	return h.service.Export(ctx, req.ID)
}

// --- Boundary class ---

// Value receiver — ok.
type Stats struct{}

func (h Stats) Handle(ctx context.Context, req *docReq) (*docResp, error) { return nil, nil }

// Only ctx and a single error result — request/response are unconstrained, ok.
type Ping struct{}

func (h Ping) Handle(ctx context.Context) error { return nil }

// Embedded interface is skipped by the dependency-naming check.
type Stream struct {
	io.Closer
}

func (h *Stream) Handle(ctx context.Context, req *docReq) (*docResp, error) { return nil, nil }
