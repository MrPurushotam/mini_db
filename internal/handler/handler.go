package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mrpurushotam/mini_database/internal/store"
)

type Handler struct {
	Store *store.Store
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{Store: store}
}

func (h *Handler) Set(c *fiber.Ctx) error {
	var kv KeyValue
	if err := c.BodyParser(&kv); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Invalid body."})
	}
	h.Store.Set(kv.Key, kv.Value)
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok"})
}

func (h *Handler) Get(c *fiber.Ctx) error {
	key := c.Query("key")
	value, exists := h.Store.Get(key)
	if !exists {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Not found"})
	}
	return c.Status(200).JSON(fiber.Map{"status": "success", "value": value})
}

func (h *Handler) GetAll(c *fiber.Ctx) error {
	key := c.Query("key")
	values := h.Store.GetAll(key)
	if len(values)!=0 {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Not found"})
	}
	return c.Status(200).JSON(fiber.Map{"status": "success", "values": values})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	key := c.Query("key")
	success := h.Store.Delete(key)
	if !success {
		return c.Status(400).JSON(fiber.Map{"status": "success", "message": "Couldn't delete key value pair."})
	}
	return c.Status(200).JSON(fiber.Map{"message": "ok", "status": "success"})
}
