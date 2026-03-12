// Package main provides the AEShield API server
//
//	@title			AEShield API
//	@version		1.0
//	@description	Secure file storage API with client-side encryption
//	@termsOfService	http://localhost:8080/terms
//
//	@contact.name	API Support
//	@contact.url	http://localhost:8080/support
//	@contact.email	support@aeshield.io
//
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/aeshield/backend/docs"
	"github.com/aeshield/backend/internal/auth"
	"github.com/aeshield/backend/internal/config"
	"github.com/aeshield/backend/internal/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.Load()

	if _, err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	userRepo := database.NewUserRepository(database.GetDB())
	if err := userRepo.CreateIndexes(context.Background()); err != nil {
		log.Printf("Warning: failed to create user indexes: %v", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	app.Static("/docs", "./docs/static")

	app.Get("/api/v1/swagger.json", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendFile("./docs/swagger.json")
	})

	authService := auth.NewService(cfg, userRepo)
	authHandler := auth.NewHandler(authService)

	api := app.Group("/api/v1")
	protected := api.Group("", auth.JWTMiddleware(cfg.JWTSecret))

	api.Get("/auth/urls", authHandler.GetAuthURLs)

	api.Get("/auth/google", authHandler.GoogleLogin)
	api.Get("/auth/google/callback", authHandler.GoogleCallback)

	api.Get("/auth/github", authHandler.GitHubLogin)
	api.Get("/auth/github/callback", authHandler.GitHubCallback)

	protected.Get("/auth/me", authHandler.Me)

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		log.Printf("Swagger available at http://localhost:%s/docs/index.html", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Disconnecting from MongoDB...")
	if err := database.GetDB().Disconnect(nil); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
	}

	log.Println("Server exited")
}
