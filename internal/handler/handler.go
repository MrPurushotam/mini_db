package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mrpurushotam/mini_database/internal/logger"
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
		logger.Error("Failed to parse Set request body", "error", err)
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Invalid body."})
	}

	h.Store.Set(kv.Key, kv.Value)
	logger.Info("Key set successfully", "key", kv.Key, "value", kv.Value)
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok"})
}

func (h *Handler) Get(c *fiber.Ctx) error {
	key := c.Query("key")
	value, exists := h.Store.Get(key)
	if !exists {
		logger.Warn("Key not found during Get operation", "key", key)
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Not found"})
	}
	logger.Info("Key retrieved successfully", "key", key)
	return c.Status(200).JSON(fiber.Map{"status": "success", "value": value})
}

func (h *Handler) GetAll(c *fiber.Ctx) error {
	values := h.Store.GetAll()
	logger.Info("Retrieved all values", "count", len(values))
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "Value fetched.", "values": values})
}

func (h *Handler) GetAllKeys(c *fiber.Ctx) error {
	keys := h.Store.GetAllKeys()
	logger.Info("Retrieved all keys", "count", len(keys))
	return c.Status(200).JSON(fiber.Map{"status": "success", "keys": keys})
}

func (h *Handler) GetAllValues(c *fiber.Ctx) error {
	values := h.Store.GetAllValues()
	logger.Info("Retrieved all values (only values)", "count", len(values))
	return c.Status(200).JSON(fiber.Map{"status": "success", "values": values})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	key := c.Query("key")
	success, err := h.Store.Delete(key)
	if err != nil {
		logger.Warn("Failed to delete key,error occurred", "key", key)
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": err})
	}
	if !success {
		logger.Warn("Failed to delete key, key not found", "key", key)
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Couldn't delete key value pair."})
	}
	logger.Info("Key deleted successfully", "key", key)
	return c.Status(200).JSON(fiber.Map{"message": "ok", "status": "success"})
}
