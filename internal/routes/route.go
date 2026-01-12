package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mrpurushotam/mini_db/internal/handler"
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

	router.Post("/SADD", func(c *fiber.Ctx) error {
		return h.SAdd(c)
	})

	router.Get("/SMEMBERS", func(c *fiber.Ctx) error {
		return h.SMembers(c)
	})

	router.Patch("/SPOP", func(c *fiber.Ctx) error {
		return h.SPop(c)
	})

	router.Post("/LPUSH", func(c *fiber.Ctx) error {
		return h.LPush(c)
	})

	router.Post("/RPUSH", func(c *fiber.Ctx) error {
		return h.RPush(c)
	})

	router.Get("/LRANGE", func(c *fiber.Ctx) error {
		return h.LRange(c)
	})

	router.Post("/ENQUEUE", func(c *fiber.Ctx) error {
		return h.Enqueue(c)
	})

	router.Patch("/DEQUEUE", func(c *fiber.Ctx) error {
		return h.Dequeue(c)
	})

	router.Post("/PUSH", func(c *fiber.Ctx) error {
		return h.Push(c)
	})

	router.Patch("/POP", func(c *fiber.Ctx) error {
		return h.Pop(c)
	})

	router.Post("/HSET", func(c *fiber.Ctx) error {
		return h.HSet(c)
	})

	router.Get("/HGET", func(c *fiber.Ctx) error {
		return h.HGet(c)
	})

	router.Get("/HGETALL", func(c *fiber.Ctx) error {
		return h.HGetAll(c)
	})

	router.Get("/snapshot", func(c *fiber.Ctx) error {
		return h.Snapshot(c)
	})

	router.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"message": "Api is running"})
	})
}
