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
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "API está funcionando!",
		})
	})

	api := app.Group("/api/v1")

	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Bem-vindo à API!",
		})
	})

	userHandler := handlers.NewUserHandler(dbpool)
	authHandler := handlers.NewAuthHandler(dbpool)
	processHandler := handlers.NewProcessHandler(dbpool)
	webhookHandler := handlers.NewWebhookHandler(dbpool)

	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)

	users := api.Group("/users")
	users.Get("/", userHandler.GetUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Post("/", userHandler.CreateUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)

	processes := api.Group("/processes")
	processes.Use(handlers.JWTMiddleware())
	processes.Post("/", processHandler.CreateProcessInfo)
	processes.Get("/", processHandler.GetProcessInfos)
	processes.Get("/:id", processHandler.GetProcessInfo)
	processes.Get("/process/:processId", processHandler.GetProcessInfosByProcessID)
	processes.Delete("/:id", processHandler.DeleteProcessInfo)

	webhook := api.Group("/webhook")
	webhook.Get("/iterate-processes", webhookHandler.IterateProcesses)
	webhook.Post("/process-by-pid", webhookHandler.ProcessByPid)
}
