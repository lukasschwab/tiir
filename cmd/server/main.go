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
	"github.com/gofiber/keyauth/v2"
	"github.com/lukasschwab/tiir/pkg/config"
	"github.com/lukasschwab/tiir/pkg/render"
	"github.com/lukasschwab/tiir/pkg/text"
)

var (
	listRenderer = map[string]render.Function{
		"json":                  render.JSON,
		"application/json":      render.JSON,
		"application/feed+json": render.JSONFeed,
		"plain":                 render.Plain,
		"text/plain":            render.Plain,
		"html":                  render.HTML,
	}
)

func main() {
	app := fiber.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	defer cfg.Service.Close()

	app.Use(logger.New())

	apiSecret := cfg.GetAPISecret()
	app.Use(keyauth.New(keyauth.Config{
		Filter:    filter(apiSecret),
		Validator: validator(apiSecret),
	}))

	// Create text.
	app.Post("/texts", func(c *fiber.Ctx) error {
		t := new(text.Text)
		if err := c.BodyParser(t); err != nil {
			c.Status(fiber.StatusBadRequest)
			return fmt.Errorf("error parsing request body: %w", err)
		}

		created, err := cfg.Service.Create(t)
		if err != nil {
			return fmt.Errorf("error writing record: %w", err)
		}
		return c.Status(fiber.StatusCreated).JSON(created)
	})

	app.Get("/texts/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			c.Status(fiber.StatusBadRequest)
			return errors.New("update request must specify record ID")
		}

		t, err := cfg.Service.Read(id)
		if err != nil {
			// BODGE: assume the text wasn't found. Makes upsert-adaptation in
			// store.http easier.
			log.Printf("error getting record: %v", err)
			c.SendStatus(fiber.StatusNotFound)
			return nil
		}
		return c.Status(fiber.StatusOK).JSON(t)
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

		updated, err := cfg.Service.Update(id, updates)
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

		deleted, err := cfg.Service.Delete(id)
		if err != nil {
			return fmt.Errorf("error deleting record: %w", err)
		}
		return c.Status(fiber.StatusOK).JSON(deleted)
	})

	// List all texts.
	app.Get("/texts", func(c *fiber.Ctx) error {
		// FIXME: handle errors.
		texts, _ := cfg.Service.List()
		// TODO: look at accept headers, not just a format parameter.
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
	if err := cfg.Service.Close(); err != nil {
		log.Printf("Error closing service: %v", err)
	}
	log.Printf("Shutdown")
}
