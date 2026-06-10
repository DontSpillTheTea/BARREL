package security

import (
	"encoding/json"
	"net/http"
	"os"
)

func RequireToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := os.Getenv("BARREL_REVIEW_TOKEN")
		if token != "" {
			reqToken := r.Header.Get("X-BARREL-REVIEW-TOKEN")
			if reqToken != token {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": "Valid X-BARREL-REVIEW-TOKEN required"})
				return
			}
		}
		next.ServeHTTP(w, r)
	}
}
