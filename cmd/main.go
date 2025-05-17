package main

import (
	"log"
	"os"

	"github.com/anton-fuji/minilink/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1/shorten", routes.ShortenURL)
}

func main() {
	app := fiber.New(fiber.Config{
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: true,
	})

	app.Use(logger.New())
	setupRoutes(app)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}
	log.Printf("Listening on %s", port)
	log.Fatal(app.Listen(port))
}
