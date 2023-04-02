package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/keyauth/v2"
)

func validator(apiSecret string) func(*fiber.Ctx, string) (bool, error) {
	return func(c *fiber.Ctx, requestKey string) (bool, error) {
		log.Printf("requestKey: %v", requestKey)

		hashedAPIKey := sha256.Sum256([]byte(apiSecret))
		hashedRequestKey := sha256.Sum256([]byte(requestKey))

		if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedRequestKey[:]) == 1 {
			return true, nil
		}
		return false, keyauth.ErrMissingOrMalformedAPIKey
	}
}

func filter(apiSecret string) func(*fiber.Ctx) bool {
	return func(c *fiber.Ctx) bool {
		if apiSecret == "" {
			log.Printf("Not checking request auth: no API secret in env")
			return true
		}
		return c.Method() == http.MethodGet
	}
}
