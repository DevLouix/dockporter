package api

import (
	"encoding/json"
	"net/http"
)

// 🛡️ ENTERPRISE GUARD 3: CORS & Preflight Handler
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow any origin (In super-strict prod, change "*" to your UI's domain)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		
		// CRITICAL: You must explicitly allow your custom headers here!
		w.Header().Set("Access-Control-Allow-Headers", "X-Auth-Token, Content-Type, X-Container-Manifest")

		// If it's a preflight request, exit successfully immediately
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware - Protects endpoints but allows OPTIONS through via CORS wrapper
func AuthMiddleware(authToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientToken := r.Header.Get("X-Auth-Token")

		// WebSocket Support: Check query param for WS connections
		if clientToken == "" {
			clientToken = r.URL.Query().Get("token")
		}

		if clientToken != authToken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized: Invalid Token"})
			return
		}

		next.ServeHTTP(w, r)
	})
}