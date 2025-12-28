package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	config "github.com/mrpurushotam/mini_database/internal"
	"github.com/mrpurushotam/mini_database/internal/aof"
	"github.com/mrpurushotam/mini_database/internal/handler"
	"github.com/mrpurushotam/mini_database/internal/logger"
	"github.com/mrpurushotam/mini_database/internal/routes"
	"github.com/mrpurushotam/mini_database/internal/store"
)

func main() {
	cfg := config.LoadConfig()
	app := fiber.New()
	app.Use(fiberLogger.New())

	logger.Init(os.Stdout, "mini_db: ", cfg.LogLevel)

	aofFile, err := aof.NewAOF(cfg.AOF_FILENAME)
	if err != nil {
		logger.Error("Failed to create AOF", "error", err)
	}
	defer aofFile.Close()

	store := store.NewStore()
	logger.Info("Store initalized")

	if err := store.LoadFromAOF(cfg.AOF_FILENAME); err != nil {
		logger.Error("Failed to load AOF", "error", err)
	}

	store.EnableAOF(aofFile)

	handler := handler.NewHandler(store)
	api := app.Group("/api/v0")
	routes.Register(api, handler)
	logger.Info("Routes registered")

	logger.Info("starting server on: ", "port", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
