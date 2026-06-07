// Eval для settings.exclude GID-163.
package handler

import "net/http"

// Исключён как "Health" — health-чеку чистый handler позволителен.
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Не исключён — репортится.
func Ready(w http.ResponseWriter, r *http.Request) { // want `GID-163: "Ready" — чистый golang handler запрещён`
	w.WriteHeader(http.StatusOK)
}
