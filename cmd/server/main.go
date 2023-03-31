package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/lukasschwab/tiir/pkg/render"
	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/tir"
)

var (
	listRenderer = map[string]render.Renderer{
		"json":             render.JSONFeed,
		"application/json": render.JSONFeed,
		"plain":            render.Plain,
		"text/plain":       render.Plain,
	}
)

func main() {
	app := fiber.New()

	// Inbound logging.
	app.Use(logger.New())

	// Create text.
	app.Post("/texts", func(c *fiber.Ctx) error {
		t := new(text.Text)
		if err := c.BodyParser(t); err != nil {
			c.Status(fiber.StatusBadRequest)
			return fmt.Errorf("error parsing request body: %w", err)
		}
		configuredService, _ := service()
		if created, err := configuredService.Create(t); err != nil {
			return fmt.Errorf("error writing record: %w", err)
		} else {
			return c.Status(fiber.StatusCreated).JSON(created)
		}
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

		configuredService, _ := service()
		if updated, err := configuredService.Update(id, updates); err != nil {
			return fmt.Errorf("error updating record: %w", err)
		} else {
			return c.Status(fiber.StatusOK).JSON(updated)
		}
	})

	// Delete text by ID.
	app.Delete("/texts/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			c.Status(fiber.StatusBadRequest)
			return errors.New("update request must specify record ID")
		}

		configuredService, _ := service()
		if deleted, err := configuredService.Delete(id); err != nil {
			return fmt.Errorf("error deleting record: %w", err)
		} else {
			return c.Status(fiber.StatusOK).JSON(deleted)
		}
	})

	// List all texts.
	app.Get("/texts", func(c *fiber.Ctx) error {
		// FIXME: handle errors.
		configuredService, _ := service()
		texts, _ := configuredService.List()

		renderer := listRenderer[c.Query("format", "plain")]
		renderer.Render(texts, c)
		return nil
	})

	log.Fatal(app.Listen(":3000"))
}

func service() (*tir.Service, error) {
	// FIXME: placeholder service.
	if home, err := os.UserHomeDir(); err != nil {
		return nil, fmt.Errorf("error getting user home directory: %v", err)
	} else if store, err := store.UseFile(home + "/.tir.json"); err != nil {
		return nil, fmt.Errorf("error opening tir file: %v", err)
	} else {
		return &tir.Service{Store: store}, nil
	}
}
