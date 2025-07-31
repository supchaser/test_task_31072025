package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/supchaser/test_task/internal/utils/logger"
	"github.com/supchaser/test_task/internal/utils/responses"
	"go.uber.org/zap"
)

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if err == http.ErrAbortHandler {
					logger.Warn("connection aborted",
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
					)
					panic(err)
				}

				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)

				responses.DoBadResponseAndLog(w, http.StatusInternalServerError, "internal server error")
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}
