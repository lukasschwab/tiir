package main

import (
	"log" // TODO: set up zap.

	"github.com/lukasschwab/tiir/pkg/edit"
	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/tir"
)

func main() {
	// TODO: parse command line arguments. This might be easier if we factor the
	// editor interface out from the data process.
	// TODO: actually persist results.
	service := tir.Service{Store: store.UseMemory()}

	// Creating a new record.
	initial := &text.Text{Title: "Your initial title here"}
	if final, err := (edit.Tea{}).Update(initial); err != nil {
		log.Fatalf("couldn't start editor: %v", err)
	} else if created, err := service.Create(final); err != nil {
		log.Fatalf("error comitting new record: %v", err)
	} else {
		log.Printf("successfully created text %v", created.ID)
	}
}
