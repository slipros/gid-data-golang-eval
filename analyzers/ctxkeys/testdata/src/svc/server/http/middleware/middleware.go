// Eval для GID-165: свой contextKey в middleware запрещён.
package middleware

import (
	"context"
	"net/http"

	"svc/domain/model"
)

type contextKey string

// --- Позитив: middleware кладёт данные своим ключом ---

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), contextKey("user"), "id") // want `GID-165: context\.WithValue вне /domain/model запрещён — ключи контекста и helper'ы живут в model, чтобы бизнес-слои не зависели от middleware`
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --- Негатив: middleware использует helper из model ---

func AuthViaModel(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := model.WithUserID(r.Context(), "id")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --- Неприменимость: производные контексты без значений ---

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(ctx)
}
