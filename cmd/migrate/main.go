package main

import (
	"log"

	"github.com/lukasschwab/tiir/pkg/config"
	"github.com/lukasschwab/tiir/pkg/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed loading configured (origin) store: %v", err)
	}
	defer cfg.App.Close()

	// NOTE: Replace this connection string.
	destinationStore, err := store.UseLibSQL("file:///Users/lukas/Programming/tiir/tempstore.db")
	if err != nil {
		log.Fatalf("Failed connecting to destination store: %v", err)
	}
	defer destinationStore.Close()

	texts, err := cfg.App.List()
	if err != nil {
		log.Fatalf("Couldn't load texts: %v", err)
	}

	log.Printf("Upserting %v texts", len(texts))
	for _, text := range texts {
		// Bodge: https://github.com/libsql/libsql-client-go/issues/79
		if name, _ := text.Timestamp.Zone(); name == "" {
			text.Timestamp = text.Timestamp.UTC()
		}

		// NOTE: these upserts should be idempotent on ID.
		if _, err := destinationStore.Upsert(text); err != nil {
			log.Printf("Error upserting text %v: %v", text.ID, err)
		}
	}
	log.Printf("Migration complete")
}
