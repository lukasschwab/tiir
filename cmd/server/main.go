package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lukasschwab/tiir/pkg/config"
	"github.com/lukasschwab/tiir/pkg/render"
	"github.com/lukasschwab/tiir/pkg/text"
)

var (
	// Bodged aliasing for List request render formats. Roughly correspond to
	// content types.
	formatRenderers = map[string]render.Function{
		"json":                  render.JSON,
		"application/json":      render.JSON,
		"application/feed+json": render.JSONFeed,
		"plain":                 render.Plain,
		"text/plain":            render.Plain,
		"html":                  render.HTML,
		"text/html":             render.HTML,
	}
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	defer cfg.App.Close()

	apiSecret := cfg.GetAPISecret()

	mux := http.NewServeMux()

	// Root redirect.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/texts", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	// List all texts.
	mux.HandleFunc("/texts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		texts, err := cfg.App.List()
		if err != nil {
			log.Printf("error listing texts: %v", err)
			http.Error(w, fmt.Sprintf("error listing texts: %v", err), http.StatusInternalServerError)
			return
		}

		// Check for format query parameter first.
		format := r.URL.Query().Get("format")
		if format != "" {
			if renderer, ok := formatRenderers[format]; ok {
				w.Header().Set("Content-Type", fmt.Sprintf("%v; charset=utf-8", format))
				if err := renderer(texts, w); err != nil {
					log.Printf("error rendering: %v", err)
				}
				return
			}
		}

		// Fall back on Accept header.
		acceptHeader := r.Header.Get("Accept")
		if acceptHeader != "" {
			for contentType := range formatRenderers {
				if strings.Contains(acceptHeader, contentType) {
					if renderer, ok := formatRenderers[contentType]; ok {
						w.Header().Set("Content-Type", fmt.Sprintf("%v; charset=utf-8", contentType))
						if err := renderer(texts, w); err != nil {
							log.Printf("error rendering: %v", err)
						}
						return
					}
				}
			}
		}

		// Fall back on HTML.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := render.HTML(texts, w); err != nil {
			log.Printf("error rendering HTML: %v", err)
		}
	})

	// Dedicated route for the JSON feed.
	mux.HandleFunc("/texts/feed.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		texts, err := cfg.App.List()
		if err != nil {
			log.Printf("error listing texts: %v", err)
			http.Error(w, fmt.Sprintf("error listing texts: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/feed+json; charset=utf-8")
		if err := render.JSONFeed(texts, w); err != nil {
			log.Printf("error rendering JSON feed: %v", err)
		}
	})

	// Create, get, update, delete text by ID (or create if no ID).
	mux.HandleFunc("/texts/", func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path /texts/:id
		path := strings.TrimPrefix(r.URL.Path, "/texts/")
		id := strings.TrimSuffix(path, "/")

		switch r.Method {
		case http.MethodPost:
			// Create text.
			t := new(text.Text)
			if err := json.NewDecoder(r.Body).Decode(t); err != nil {
				http.Error(w, fmt.Sprintf("error parsing request body: %v", err), http.StatusBadRequest)
				return
			}

			created, err := cfg.App.Create(t)
			if err != nil {
				log.Printf("error writing record: %v", err)
				http.Error(w, fmt.Sprintf("error writing record: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(created); err != nil {
				log.Printf("error encoding response: %v", err)
			}

		case http.MethodGet:
			// Get text by ID.
			if id == "" {
				http.Error(w, "request must specify record ID", http.StatusBadRequest)
				return
			}

			t, err := cfg.App.Read(id)
			if err != nil {
				// BODGE: assume the text wasn't found. Makes upsert-adaptation in
				// store.http easier.
				log.Printf("error getting record: %v", err)
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(t); err != nil {
				log.Printf("error encoding response: %v", err)
			}

		case http.MethodPatch:
			// Update text by ID.
			if id == "" {
				http.Error(w, "update request must specify record ID", http.StatusBadRequest)
				return
			}

			updates := new(text.Text)
			if err := json.NewDecoder(r.Body).Decode(updates); err != nil {
				http.Error(w, fmt.Sprintf("error parsing request body: %v", err), http.StatusBadRequest)
				return
			}

			updated, err := cfg.App.Update(id, updates)
			if err != nil {
				log.Printf("error updating record: %v", err)
				http.Error(w, fmt.Sprintf("error updating record: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(updated); err != nil {
				log.Printf("error encoding response: %v", err)
			}

		case http.MethodDelete:
			// Delete text by ID.
			if id == "" {
				http.Error(w, "delete request must specify record ID", http.StatusBadRequest)
				return
			}

			deleted, err := cfg.App.Delete(id)
			if err != nil {
				log.Printf("error deleting record: %v", err)
				http.Error(w, fmt.Sprintf("error deleting record: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(deleted); err != nil {
				log.Printf("error encoding response: %v", err)
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Static file serving.
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))

	// Wrap with middleware.
	handler := loggingMiddleware(authMiddleware(apiSecret, mux))

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Start server in a goroutine.
	go func() {
		log.Printf("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("shutting down: %v", err)
		}
	}()

	// Wait for interrupt signal.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // Block main thread until interrupt.
	log.Printf("Gracefully shutting down...")

	// Graceful shutdown with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if err := cfg.App.Close(); err != nil {
		log.Printf("Error closing service: %v", err)
	}
	log.Printf("Shutdown")
}
