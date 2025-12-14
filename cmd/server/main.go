package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	config "github.com/mrpurushotam/mini_database/internal"
	"github.com/mrpurushotam/mini_database/internal/handler"
	"github.com/mrpurushotam/mini_database/internal/logger"
	"github.com/mrpurushotam/mini_database/internal/routes"
	"github.com/mrpurushotam/mini_database/internal/store"
)

func main() {

	cfg := config.LoadConfig()
	app := fiber.New()
	app.Use(fiberLogger.New())

	appLogger := logger.NewStdLogger(os.Stdout, "mini_db: ", cfg.LogLevel)

	store := store.NewStore(appLogger)
	appLogger.Info("Store initalized", "keys", len(store.GetAllKeys()))

	handler := handler.NewHandler(store, appLogger)
	api := app.Group("/api/v0")
	routes.Register(api, handler, appLogger)
	appLogger.Info("Routes registered")

	appLogger.Info("starting server on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
