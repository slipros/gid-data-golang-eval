// Not applicable: outside server/http the rule does not apply.
package middleware

import "net/http"

func Logger(w http.ResponseWriter, r *http.Request) {}
