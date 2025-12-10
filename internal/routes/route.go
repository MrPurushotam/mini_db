package routes

import "github.com/gofiber/fiber/v2"

func Register(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"message": "Api is running"})
	})
}
