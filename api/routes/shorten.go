package routes

import (
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/todzuko/url-shortener/database"
	"github.com/todzuko/url-shortener/helpers"
	"os"
	"strconv"
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

	err := checkAPILimit(c)
	if err != nil {
		return err
	}
	body.URL = helpers.EnforceHTTP(body.URL)
	decrAPILimit(c)
	return nil
}

// checkAPILimit checks if user has entries left
func checkAPILimit(c *fiber.Ctx) error {
	r2 := database.CreateClient(1)
	defer r2.Close()

	count, err := r2.Get(database.Ctx, c.IP()).Result()
	if err != redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_LIMIT"), time.Second*3600).Err()
	} else {
		intCount, _ := strconv.Atoi(count)
		if intCount <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":            "Rate limit exceeded",
				"rate-limit-reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}
	return nil
}

// decrAPILimit decreases Limit Counter for user API
func decrAPILimit(c *fiber.Ctx) {
	r2 := database.CreateClient(1)
	r2.Decr(database.Ctx, c.IP())
}
