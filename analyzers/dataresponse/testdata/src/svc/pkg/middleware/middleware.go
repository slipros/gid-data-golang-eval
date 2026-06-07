// Неприменимость: вне server/http правило не действует.
package middleware

import "net/http"

func Logger(w http.ResponseWriter, r *http.Request) {}
