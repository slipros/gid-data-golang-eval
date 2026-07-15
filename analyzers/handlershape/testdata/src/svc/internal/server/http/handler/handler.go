// Eval GID-230: non-applicability — an HTTP handler package
// (internal/server/http/handler) is out of scope. Its handlers follow the
// data-response.go shape Handle(*http.Request, *dataresponse.Factory)
// *response.DataResponse and are governed by GID-162/GID-163, so neither the
// gRPC Handle shape nor the <Handler>Validator/<Handler>Service naming applies.
// No diagnostic must be reported here.
package handler

import "net/http"

// Local stand-ins for github.com/raoptimus/data-response.go/v2 types, so the
// fixture stays dependency-free while keeping the real handler shape.
type dataresponseFactory struct{}
type responseDataResponse struct{}

// uploadService is an interface dependency NOT named after the handler; under
// the old grpc-only-shaped rule this field alone triggered GID-230.
type uploadService interface {
	Upload(r *http.Request) error
}

// UploadTranscribeJobAudio is a canonical data-response.go HTTP handler: its
// Handle takes *http.Request (not context.Context) and returns
// *responseDataResponse (not error). It must NOT be flagged.
type UploadTranscribeJobAudio struct {
	service uploadService
}

func (h *UploadTranscribeJobAudio) Handle(r *http.Request, f *dataresponseFactory) *responseDataResponse {
	if err := h.service.Upload(r); err != nil {
		return &responseDataResponse{}
	}
	return &responseDataResponse{}
}
