// Eval для GID-163 (data-response.go вместо чистых хендлеров).
package handler

import "net/http"

type Snapshot struct{}

// --- Позитив: чистый golang handler ---

func (h *Snapshot) Get(w http.ResponseWriter, r *http.Request) { // want `GID-163: "Get" — чистый golang handler запрещён, используйте github\.com/raoptimus/data-response\.go/v2 \(исключения: nolint или settings\.exclude\)`
	w.WriteHeader(http.StatusOK)
}

// Граничный кейс: package-level чистый handler.
func List(w http.ResponseWriter, r *http.Request) { // want `GID-163: "List" — чистый golang handler запрещён`
	w.WriteHeader(http.StatusOK)
}

// --- Негатив: иная сигнатура — не чистый handler ---

func (h *Snapshot) Convert(r *http.Request) string { return r.URL.Path }

// Неприменимость: ResponseWriter без Request.
func write(w http.ResponseWriter, body []byte) {
	_, _ = w.Write(body) //nolint // errcheck вне этого eval
}
