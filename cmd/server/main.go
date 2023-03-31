package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/lukasschwab/tiir/pkg/render"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/tir"
)

var (
	listRenderer = map[string]render.Function{
		"json":             render.JSONFeed,
		"application/json": render.JSONFeed,
		"plain":            render.Plain,
		"text/plain":       render.Plain,
		"html":             render.HTML,
	}
)

func main() {
	app := fiber.New()

	configuredService, _ := tir.FromConfig()
	defer configuredService.Close()

	// Inbound logging.
	app.Use(logger.New())

	// TODO: think about simple auth.

	// Create text.
	app.Post("/texts", func(c *fiber.Ctx) error {
		t := new(text.Text)
		if err := c.BodyParser(t); err != nil {
			c.Status(fiber.StatusBadRequest)
			return fmt.Errorf("error parsing request body: %w", err)
		}

		created, err := configuredService.Create(t)
		if err != nil {
			return fmt.Errorf("error writing record: %w", err)
		}
		return c.Status(fiber.StatusCreated).JSON(created)
	})

	// Update text by ID.
	app.Patch("/texts/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			c.Status(fiber.StatusBadRequest)
			return errors.New("update request must specify record ID")
		}

		updates := new(text.Text)
		if err := c.BodyParser(updates); err != nil {
			c.Status(fiber.StatusBadRequest)
			return fmt.Errorf("error parsing request body: %w", err)
		}

		updated, err := configuredService.Update(id, updates)
		if err != nil {
			return fmt.Errorf("error updating record: %w", err)
		}
		return c.Status(fiber.StatusOK).JSON(updated)
	})

	// Delete text by ID.
	app.Delete("/texts/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			c.Status(fiber.StatusBadRequest)
			return errors.New("update request must specify record ID")
		}

		deleted, err := configuredService.Delete(id)
		if err != nil {
			return fmt.Errorf("error deleting record: %w", err)
		}
		return c.Status(fiber.StatusOK).JSON(deleted)
	})

	// List all texts.
	app.Get("/texts", func(c *fiber.Ctx) error {
		// FIXME: handle errors.
		texts, _ := configuredService.List()
		renderer := listRenderer[c.Query("format", "html")]
		c.Set("Content-Type", "text/html; charset=utf-8")
		return renderer(texts, c)
	})

	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Fatalf("shutting down: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // Block main thread until interrupt.
	log.Printf("Gracefully shutting down...")
	_ = app.ShutdownWithTimeout(5 * time.Second)
	if err := configuredService.Close(); err != nil {
		log.Printf("Error closing service: %v", err)
	}
	log.Printf("Shutdown")
}
