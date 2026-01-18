package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	config "github.com/mrpurushotam/mini_db/internal"
	"github.com/mrpurushotam/mini_db/internal/aof"
	"github.com/mrpurushotam/mini_db/internal/handler"
	"github.com/mrpurushotam/mini_db/internal/logger"
	"github.com/mrpurushotam/mini_db/internal/routes"
	"github.com/mrpurushotam/mini_db/internal/store"
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
	defer func() {
		if aofFile != nil {
			aofFile.Close()
		}
	}()

	store := store.NewStore()
	logger.Info("Store initialized")
	store.EnableAOF(aofFile)

	if err := store.LoadFromAOF(cfg.AOF_FILENAME); err != nil {
		logger.Error("Failed to load AOF", "error", err)
	}

	handler := handler.NewHandler(store)
	api := app.Group("/api/v0")
	routes.Register(api, handler)
	logger.Info("Routes registered")

	var wg sync.WaitGroup

	if aofFile == nil {
		logger.Info("AOF not initialized; auto-snapshot disabled")
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// initial snapshot at startup
			if err := aofFile.Snapshot(store); err != nil {
				logger.Error("Initial snapshot failed", "error", err)
			} else {
				logger.Info("Initial snapshot completed")
			}

			ticker := time.NewTicker(6 * time.Hour)
			defer ticker.Stop()

			for range ticker.C {
				if err := aofFile.Snapshot(store); err != nil {
					logger.Error("Auto-snapshot failed", "error", err)
				} else {
					logger.Info("Auto-snapshot completed")
				}
			}
		}()
	}
	logger.Info("starting server", "port", cfg.Port)

	if cfg.AWS_LAMBDA_FUNCTION_NAME != "" {
		adapter := fiberadapter.New(app)
		lambda.Start(adapter.ProxyWithContext)
		return
	}

	log.Fatal(app.Listen(":" + cfg.Port))
}
