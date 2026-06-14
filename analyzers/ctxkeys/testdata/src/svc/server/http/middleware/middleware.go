// Eval for GID-165: a custom contextKey in a middleware is forbidden.
package middleware

import (
	"context"
	"net/http"

	"svc/domain/model"
)

type contextKey string

// --- Positive: the middleware stores data with its own key ---

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), contextKey("user"), "id") // want `GID-165: context\.WithValue outside /domain/model is forbidden\. Fix: keep context keys and helpers in /domain/model so business layers do not depend on middleware`
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --- Negative: the middleware uses a helper from model ---

func AuthViaModel(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := model.WithUserID(r.Context(), "id")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --- Not applicable: derived contexts without values ---

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(ctx)
}
