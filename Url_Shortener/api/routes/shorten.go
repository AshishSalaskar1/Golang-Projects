package routes

import (
	"os"
	"strconv"
	"time"
	"url-shortener-redis-fiber/database"
	"url-shortener-redis-fiber/helpers"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := request{}

	// you are passing ptr to `body`, so that gets updated with data
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// implement rate limiting: You want to allow user to shorten at max 10 urls in 30 minutes
	rdbRateLimiter := database.CreateClient(0)
	defer rdbRateLimiter.Close()

	val, err := rdbRateLimiter.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		// set IP: API_QUOTA (which auto expires in 30 minutes)
		_ = rdbRateLimiter.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else if err == nil {
		valInt, _ := strconv.Atoi(val)

		if valInt <= 0 {
			// Get the time remaining until the reset time
			limit, _ := rdbRateLimiter.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Rate limit exceeded. Limit resets after " + limit.String()})
		}
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot connect to DB"})
	}

	// check if the url is valid
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	// check for domain error (invalid domain)
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Domain error, please provide a valid URL"})
	}

	// enforce https, SSL
	body.URL = helpers.EnforceHTTPS(body.URL)

	// Shorten the URL
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	// instance of ID <-> URL mapping
	rdbMain := database.CreateClient(0)
	defer rdbMain.Close()

	// check if id is alreay being used
	val, _ = rdbMain.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "URL custom short is already in use"})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = rdbMain.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to connect to DB"})
	}

	// response
	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10, // API_QUOTA
		XRateLimitReset: 30,
	}

	// decrement the API quota
	rdbRateLimiter.Decr(database.Ctx, c.IP())

	// return response
	val, _ = rdbRateLimiter.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := rdbRateLimiter.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Hour / 30

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
	return c.Status(fiber.StatusOK).JSON(resp)

}
