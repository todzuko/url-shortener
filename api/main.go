package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/todzuko/url-shortener/routes"
	"log"
	"os"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ShortenUrl)
	app.Post("/:url", routes.ResolveUrl)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	app := fiber.New()

	setupRoutes(app)
	log.Fatal(app.Listen(os.Getenv("APP_DOMAIN")))
}
