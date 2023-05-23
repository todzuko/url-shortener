package routes

import (
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	err := validateURL(c, *body)
	if err != nil {
		return err
	}

	err = checkAPILimit(c)
	if err != nil {
		return err
	}

	body.URL = helpers.EnforceHTTP(body.URL)

	var id string
	if body.Short == "" {
		id = uuid.New().String()
	} else {
		id = body.Short
	}
	r := database.CreateClient(0)
	defer r.Close()

	val, _ := r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Provided short url is already in use"})
	}

	if body.Exp == 0 {
		body.Exp = 24
	}
	//set full url to sort
	err = r.Set(database.Ctx, id, body.URL, body.Exp*3600*time.Second).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to connect to server"})
	}

	res := response{
		URL:             body.URL,
		Short:           "",
		Exp:             body.Exp,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	decrAPILimit(c)
	r2 := database.CreateClient(1)
	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	res.XRateRemaining, _ = strconv.Atoi(val)
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	res.XRateLimitReset = ttl / time.Nanosecond / time.Minute
	res.Short = os.Getenv("DOMAIN") + "/" + id
	return c.Status(fiber.StatusOK).JSON(res)
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

func validateURL(c *fiber.Ctx, body request) error {
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Service error"})
	}
	return nil
}

// decrAPILimit decreases Limit Counter for user API
func decrAPILimit(c *fiber.Ctx) {
	r2 := database.CreateClient(1)
	r2.Decr(database.Ctx, c.IP())
}
