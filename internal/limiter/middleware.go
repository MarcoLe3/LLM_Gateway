package limiter

import (
	"llm_gateway/internal/auth"
	"log/slog"
	"net/http"
	"strconv"
)

func (l *Limiter) Middleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		caller, ok := auth.GetCallerFromContext(r.Context())
		if !ok {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		res, err := l.Allow(r.Context(), caller.ID, caller.RateLimitRPM)
		if err != nil {
			slog.Error("rate limiter error", "err", err)
		}

		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		if !res.Allowed() {
			w.Header().Set("Retry-After", strconv.Itoa(int(math)))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
