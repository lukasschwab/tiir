package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	// acceptPrecedence defines a deterministic order for Accept header
	// negotiation, checked against the formatRenderers map.
	acceptPrecedence = []string{
		"application/json",
		"application/feed+json",
		"text/plain",
		"text/html",
	}
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	defer func() {
		if err := cfg.App.Close(); err != nil {
			log.Printf("Error closing service: %v", err)
		}
	}()

	apiSecret := cfg.GetAPISecret()

	mux := http.NewServeMux()

	// Root redirect.
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/texts", http.StatusFound)
	})

	// List all texts.
	mux.HandleFunc("GET /texts", func(w http.ResponseWriter, r *http.Request) {
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
			for _, contentType := range acceptPrecedence {
				if strings.Contains(acceptHeader, contentType) {
					renderer := formatRenderers[contentType]
					w.Header().Set("Content-Type", fmt.Sprintf("%v; charset=utf-8", contentType))
					if err := renderer(texts, w); err != nil {
						log.Printf("error rendering: %v", err)
					}
					return
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
	mux.HandleFunc("GET /texts/feed.json", func(w http.ResponseWriter, r *http.Request) {
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

	// Create text.
	mux.HandleFunc("POST /texts", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// Get text by ID.
	mux.HandleFunc("GET /texts/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

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
	})

	// Update text by ID.
	mux.HandleFunc("PATCH /texts/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

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
	})

	// Delete text by ID.
	mux.HandleFunc("DELETE /texts/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

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

	// ctx is canceled on SIGINT/SIGTERM; stop() also cancels it, which we
	// use to unify signal-driven and error-driven shutdown paths.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Printf("Gracefully shutting down: %v", context.Cause(ctx))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Printf("Shutdown")
}
