package middleware

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

func Rate(next http.HandlerFunc, rps int) http.HandlerFunc {
	burst := max(rps/10, 5)
	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if err := limiter.Wait(ctx); err != nil {
			http.Error(w, "Rate limit timeout", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
