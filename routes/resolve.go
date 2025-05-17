package routes

import (
	"context"

	"github.com/anton-fuji/minilink/databases"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func resolveURL(c *fiber.Ctx) error {
	ctx := context.Background()
	url := c.Params("url")
	r := databases.CreateClient(0)
	defer r.Close()

	val, err := r.Get(databases.Ctx, url).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "short not found in database",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot connect to DB",
		})
	}

	// アクセスカウンター
	rInr := databases.CreateClient(1)
	defer rInr.Close()

	_, err = rInr.Incr(ctx, "counter").Result()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot update access counter",
		})
	}

	return c.Redirect(val, fiber.StatusMovedPermanently)
}
