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

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func RegisterAuthRoutes() {
	http.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		expectedUser := os.Getenv("BARREL_DEMO_USERNAME")
		if expectedUser == "" {
			expectedUser = "evaluator"
		}
		expectedPass := os.Getenv("BARREL_DEMO_PASSWORD")
		if expectedPass == "" || req.Username != expectedUser || req.Password != expectedPass {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		token := os.Getenv("BARREL_REVIEW_TOKEN")
		if token == "" {
			token = "demo-eval-token"
		}

		json.NewEncoder(w).Encode(map[string]string{"token": token})
	})

	http.HandleFunc("/api/v1/auth/me", RequireToken(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"user": "evaluator"})
	}))

	http.HandleFunc("/api/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
	})
}
