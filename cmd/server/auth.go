package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"
	"strings"
	"time"
)

// authMiddleware provides API key authentication for non-GET requests.
func authMiddleware(apiSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if no API secret is configured.
		if apiSecret == "" {
			log.Printf("Not checking request auth: no API secret in env")
			next.ServeHTTP(w, r)
			return
		}

		// Skip authentication for GET requests.
		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Extract API key from Authorization header.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing or malformed API Key", http.StatusUnauthorized)
			return
		}

		// Support "Bearer <token>" format.
		requestKey := strings.TrimPrefix(authHeader, "Bearer ")
		requestKey = strings.TrimSpace(requestKey)

		log.Printf("requestKey: %v", requestKey)

		// Compare hashes using constant-time comparison.
		hashedAPIKey := sha256.Sum256([]byte(apiSecret))
		hashedRequestKey := sha256.Sum256([]byte(requestKey))

		if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedRequestKey[:]) != 1 {
			http.Error(w, "Invalid or missing API Key", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("%s %s - completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
}
