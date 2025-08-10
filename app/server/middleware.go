package server

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/vladazn/danish/common/userid"
)

// Helper to capture response status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			logger.Debug("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.statusCode),
				zap.String("remote", r.RemoteAddr),
				zap.Duration("duration", duration),
			)
		})
	}
}

func firebaseAuthMiddleware(fa *FirebaseClient, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
				return
			}

			idToken := strings.TrimPrefix(authHeader, "Bearer ")
			ctx := r.Context()

			authClient, err := fa.app.Auth(ctx)
			if err != nil {
				http.Error(w, "Failed to initialize Firebase Auth", http.StatusInternalServerError)
				return
			}

			token, err := authClient.VerifyIDToken(ctx, idToken)
			if err != nil {
				log.Error("firebase auth error on validate", zap.Error(err))
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			uid := token.UID
			ctx = userid.ToCtx(ctx, uid)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
