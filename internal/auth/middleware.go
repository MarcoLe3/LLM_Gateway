package auth

import (
	"net/http"
	"strings"
	"context"
)

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if raw == "" {
			http.Error(w, "missing api key", http.StatusUnauthorized)
			return 
		}

		caller, err := a.Authenticate(r.Context(), raw)
		if err != nil{
			http.Error(w, "invalid api key", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(),callerKey{}, caller)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}