package main

import (
	"context"
	"log"
	"os"

	"go-api/internal/config"
	"go-api/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbpool.Close()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(cors.New())

	setupRoutes(app, dbpool)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Servidor rodando na porta %s", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App, dbpool *pgxpool.Pool) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "API is running",
		})
	})

	// API v1
	api := app.Group("/api/v1")

	// Auth routes (no JWT required)
	authHandler := handlers.NewAuthHandler(dbpool)
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)

	// User routes (no JWT required for now)
	userHandler := handlers.NewUserHandler(dbpool)
	users := api.Group("/users")
	users.Get("/", userHandler.GetUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Post("/", userHandler.CreateUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)

	// Process routes (JWT required)
	processHandler := handlers.NewProcessHandler(dbpool)
	processes := api.Group("/processes")
	processes.Use(handlers.JWTMiddleware())

	// Snapshot routes
	processes.Get("/snapshots", processHandler.GetSnapshots)
	processes.Get("/snapshots/type/:type", processHandler.GetSnapshotsByType)
	processes.Get("/snapshots/:id", processHandler.GetSnapshot)
	processes.Get("/snapshots/:id/processes", processHandler.GetSnapshotProcesses)
	processes.Get("/snapshots/:id/queries", processHandler.GetSnapshotQueries)
	processes.Delete("/snapshots/:id", processHandler.DeleteSnapshot)

	// Process info routes
	processes.Get("/", processHandler.GetProcessInfos)
	processes.Get("/:id", processHandler.GetProcessInfo)
	processes.Get("/pid/:pid", processHandler.GetProcessInfosByProcessID)
	processes.Delete("/:id", processHandler.DeleteProcessInfo)

	// Query history and statistics
	processes.Get("/queries/history", processHandler.GetQueryHistory)
	processes.Get("/statistics", processHandler.GetStatistics)

	// Webhook routes (no JWT required, but can use JWT if provided)
	webhookHandler := handlers.NewWebhookHandler(dbpool)
	webhook := api.Group("/webhook")
	webhook.Post("/iterate-processes", webhookHandler.IterateProcesses)
	webhook.Post("/process-by-pid", webhookHandler.ProcessByPid)
}
