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

type SetKeyValue struct {
	Key     string   `json:"key"`
	Members []string `json:"value"`
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

func (h *Handler) SAdd(c *fiber.Ctx) error {
	var req SetKeyValue
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse SADD request body")
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Invalid Body"})
	}

	if req.Key == "" || len(req.Members) == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Key and value are required"})
	}

	if err := h.Store.SAdd(req.Key, req.Members...); err != nil {
		logger.Warn("SADD failed", "key", req.Key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("SADD success", "key", req.Key, "count", len(req.Members))
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok"})
}

func (h *Handler) SMembers(c *fiber.Ctx) error {
	key := c.Query("key")
	members, err := h.Store.SMembers(key)
	if err != nil {
		logger.Warn("SMEMBERS failed", "key", key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("SMEMBERS success", "key", key)
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok", "members": members})
}

func (h *Handler) SPop(c *fiber.Ctx) error {
	var req SetKeyValue
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse SPOP request body")
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Invalid Body"})
	}

	if req.Key == "" || len(req.Members) == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Key and value are required"})
	}

	members, err := h.Store.SPop(req.Key, req.Members...)
	if err != nil {
		logger.Warn("SPOP failed", "key", req.Key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("SPOP success", "key", req.Key)
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok", "members": members})
}

// --- List Operations ---

type ListKeyValue struct {
	Key   string   `json:"key"`
	Value []string `json:"value"`
}

func (h *Handler) LPush(c *fiber.Ctx) error {
	var req ListKeyValue
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse LPUSH request body")
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Invalid Body"})
	}

	if req.Key == "" || len(req.Value) == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Key and value are required"})
	}

	if err := h.Store.LPush(req.Key, req.Value...); err != nil {
		logger.Warn("LPUSH failed", "key", req.Key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("LPUSH success", "key", req.Key, "count", len(req.Value))
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok"})
}

func (h *Handler) RPush(c *fiber.Ctx) error {
	var req ListKeyValue
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse RPUSH request body")
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Invalid Body"})
	}

	if req.Key == "" || len(req.Value) == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Key and value are required"})
	}

	if err := h.Store.RPush(req.Key, req.Value...); err != nil {
		logger.Warn("RPUSH failed", "key", req.Key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("RPUSH success", "key", req.Key, "count", len(req.Value))
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok"})
}

func (h *Handler) LRange(c *fiber.Ctx) error {
	key := c.Query("key")
	start := c.QueryInt("start", 0)
	stop := c.QueryInt("stop", -1)

	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Key is required"})
	}

	values, err := h.Store.LRange(key, start, stop)
	if err != nil {
		logger.Warn("LRANGE failed", "key", key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("LRANGE success", "key", key, "start", start, "stop", stop)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "ok", "values": values})
}

// --- Queue Operations ---

func (h *Handler) Enqueue(c *fiber.Ctx) error {
	var req KeyValue
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse Enqueue request body")
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Invalid Body"})
	}

	if req.Key == "" || len(req.Value) == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Key and value are required"})
	}

	if err := h.Store.Enqueue(req.Key, req.Value); err != nil {
		logger.Warn("ENQUEUE failed", "key", req.Key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("ENQUEUE success", "key", req.Key, "value", req.Value)
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok"})
}

func (h *Handler) Dequeue(c *fiber.Ctx) error {
	key := c.Query("key")
	if key == "" {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Key is required"})
	}
	val, err := h.Store.Dequeue(key)
	if err != nil {
		logger.Warn("DEQUEUE failed", "key", key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("DEQUEUE success", "key", key, "value", val)
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok", "value": val})
}

// --- Stack Operations ---

func (h *Handler) Push(c *fiber.Ctx) error {
	var req KeyValue
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse STACK PUSH request body")
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Invalid Body"})
	}

	if req.Key == "" || len(req.Value) == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Key and value are required"})
	}

	if err := h.Store.Push(req.Key, req.Value); err != nil {
		logger.Warn("STACK PUSH failed", "key", req.Key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("STACK PUSH success", "key", req.Key, "value", req.Value)
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok"})
}

func (h *Handler) Pop(c *fiber.Ctx) error {
	key := c.Query("key")
	if key == "" {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Key is required"})
	}
	val, err := h.Store.Dequeue(key)
	if err != nil {
		logger.Warn("STACK POP failed", "key", key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("STACK POP success", "key", key, "value", val)
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "ok", "value": val})
}

// --- HashSet Operations ---

type HashmapKeyValue struct {
	Key   string `json:"key"`
	Field string `json:"field"`
	Value string `json:"value"`
}

func (h *Handler) HSet(c *fiber.Ctx) error {
	var req HashmapKeyValue
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse HSET request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid body"})
	}
	if req.Key == "" || req.Field == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "key and field are required"})
	}

	if err := h.Store.HSet(req.Key, req.Field, req.Value); err != nil {
		logger.Warn("HSET failed", "key", req.Key, "field", req.Field, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("HSET success", "key", req.Key, "field", req.Field)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "ok"})
}

func (h *Handler) HGet(c *fiber.Ctx) error {
	key := c.Query("key")
	field := c.Query("field")
	if key == "" || field == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "key and field are required"})
	}

	val, err := h.Store.HGet(key, field)
	if err != nil {
		logger.Warn("HGET failed", "key", key, "field", field, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("HGET success", "key", key, "field", field)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "value": val})
}

func (h *Handler) HGetAll(c *fiber.Ctx) error {
	key := c.Query("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "key is required"})
	}

	m, err := h.Store.HGetAll(key)
	if err != nil {
		logger.Warn("HGETALL failed", "key", key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	logger.Info("HGETALL success", "key", key, "count", len(m))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "map": m})
}
