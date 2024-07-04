package main

import (
	"fmt"
	"log"
	"os"
	"url-shortener-redis-fiber/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	app := fiber.New()
	app.Use(logger.New()) // Logger middleware

	setupRoutes(app)

	log.Fatal(app.Listen(os.Getenv("APP_PORT")))

}
