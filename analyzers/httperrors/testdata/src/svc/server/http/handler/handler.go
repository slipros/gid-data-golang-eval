// Eval для GID-162 (http handler обрабатывает ошибки сам).
package handler

import "net/http"

type Snapshot struct{}

// --- Позитив: супер-метод обработки ошибок ---

func (h *Snapshot) handleError(w http.ResponseWriter, err error) { // want `GID-162: "handleError" — супер-метод обработки ошибок запрещён, http handler обрабатывает ошибки внутри себя`
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// Граничный кейс: package-level супер-функция.
func writeError(w http.ResponseWriter, status int, err error) { // want `GID-162: "writeError" — супер-метод обработки ошибок запрещён`
	http.Error(w, err.Error(), status)
}

// --- Позитив: handler возвращает error наружу ---

func (h *Snapshot) Get(w http.ResponseWriter, r *http.Request) error { // want `GID-162: http handler "Get" не возвращает error — ошибка обрабатывается на месте`
	return nil
}

// --- Негатив: handler обрабатывает ошибку внутри ---

func (h *Snapshot) List(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// --- Неприменимость: функции без ResponseWriter ---

func convert(err error) string { return err.Error() }
