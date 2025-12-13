package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mrpurushotam/mini_database/internal/handler"
)

func Register(router fiber.Router, h *handler.Handler) {
	router.Post("/set", func(c *fiber.Ctx) error {
		return h.Set(c)
	})

	router.Get("/get", func(c *fiber.Ctx) error {
		return h.Get(c)
	})

	router.Delete("/delete", func(c *fiber.Ctx) error {
		return h.Delete(c)
	})

	router.Get("/get/all", func(c *fiber.Ctx) error {
		return h.GetAll(c)
	})

	router.Get("/keys/all", func(c *fiber.Ctx) error {
		return h.GetAllKeys(c)
	})

	router.Get("/values/all", func(c *fiber.Ctx) error {
		return h.GetAllValues(c)
	})

	router.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"message": "Api is running"})
	})
}
