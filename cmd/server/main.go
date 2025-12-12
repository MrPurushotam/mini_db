package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	config "github.com/mrpurushotam/mini_database/internal"
	"github.com/mrpurushotam/mini_database/internal/handler"
	"github.com/mrpurushotam/mini_database/internal/routes"
	"github.com/mrpurushotam/mini_database/internal/store"
)

func main() {

	cfg := config.LoadConfig()
	app := fiber.New()
	store := store.NewStore()

	handler := handler.NewHandler(store)
	api := app.Group("/api/v0")
	routes.Register(api, handler)

	log.Fatal(app.Listen(":" + cfg.Port))
}
