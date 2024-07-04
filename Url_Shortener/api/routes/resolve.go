package routes

import (
	"log"
	"url-shortener-redis-fiber/database"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx) error {

	url := c.Params("url") //:url, inline params
	log.Println(url)

	rc := database.CreateClient(0)
	defer rc.Close() // close reddis client at last

	value, err := rc.Get(database.Ctx, url).Result() // Get key from Redis

	if err == redis.Nil { // if key is not found
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Short URL not found in the database"})
	} else if err != nil { // if redis error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot connect to DB"})
	}

	// decrement the rInc counter
	rInr := database.CreateClient(1)
	defer rInr.Close()
	_ = rInr.Incr(database.Ctx, "counter") // increment value of counter by 1

	// redirect to url

	return c.Redirect(value, 301)

}
