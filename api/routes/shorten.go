package routes

import (
	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/todzuko/url-shortener/helpers"
	"time"
)

type request struct {
	URL   string        `json:"url"`
	Short string        `json:"short"`
	Exp   time.Duration `json:"exp"`
}

type response struct {
	URL             string        `json:"url"`
	Short           string        `json:"short"`
	Exp             time.Duration `json:"exp"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenUrl(c *fiber.Ctx) error {
	body := new(request)
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Service error"})
	}

	body.URL = helpers.EnforceHTTP(body.URL)

	return nil
}
