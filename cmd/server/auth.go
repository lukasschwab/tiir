package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/keyauth/v2"
)

var (
	// TODO: move server into a CLI command to share flag parsing? Currently
	// there's partial overlap, which makes running a server and a client on the
	// same machine a little awkward.
	apiKey = os.Getenv("TIR_API_SECRET")
)

func validateAPIKey(c *fiber.Ctx, requestKey string) (bool, error) {
	log.Printf("requestKey: %v", requestKey)

	hashedApiKey := sha256.Sum256([]byte(apiKey))
	hashedRequestKey := sha256.Sum256([]byte(requestKey))

	if subtle.ConstantTimeCompare(hashedApiKey[:], hashedRequestKey[:]) == 1 {
		return true, nil
	}
	return false, keyauth.ErrMissingOrMalformedAPIKey
}

func authFilter(c *fiber.Ctx) bool {
	if apiKey == "" {
		log.Printf("Danger: not checking request auth: no secret in environment")
		return true
	}
	return c.Method() == http.MethodGet
}
