package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	_ "github.com/aeshield/backend/docs"
	"github.com/aeshield/backend/internal/auth"
	"github.com/aeshield/backend/internal/config"
	"github.com/aeshield/backend/internal/database"
	"github.com/aeshield/backend/internal/files"
	"github.com/aeshield/backend/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.Load()

	// Get project root directory
	// Ưu tiên env var FRONTEND_DIST nếu được set
	frontendDist := os.Getenv("FRONTEND_DIST")
	if frontendDist == "" {
		// Fallback: dùng working directory (go run chạy từ thư mục backend)
		wd, _ := os.Getwd()
		// Nếu đang ở trong thư mục backend, lên 1 cấp
		if filepath.Base(wd) == "backend" {
			wd = filepath.Dir(wd)
		}
		frontendDist = filepath.Join(wd, "frontend", "dist")
	}

	if _, err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	userRepo := database.NewUserRepository(database.GetDB())
	if err := userRepo.MigrateOldSchema(context.Background()); err != nil {
		log.Printf("Warning: failed to migrate old schema: %v", err)
	}
	if err := userRepo.CreateIndexes(context.Background()); err != nil {
		log.Printf("Warning: failed to create user indexes: %v", err)
	}

	// Setup R2 và file storage
	r2Client, err := storage.NewR2Client(cfg)
	if err != nil {
		log.Fatalf("Failed to init R2 client: %v", err)
	}
	fileRepo := storage.NewFileRepository(database.GetDB().Database)
	if err := fileRepo.CreateIndexes(context.Background()); err != nil {
		log.Printf("Warning: failed to create file indexes: %v", err)
	}
	fileService := files.NewService(r2Client, fileRepo)
	fileHandler := files.NewHandler(fileService)

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

	// Public routes
	app.Get("/api/v1/auth/urls", authHandler.GetAuthURLs)
	app.Get("/api/v1/auth/google", authHandler.GoogleLogin)
	app.Get("/api/v1/auth/google/callback", authHandler.GoogleCallback)
	app.Get("/api/v1/auth/github", authHandler.GitHubLogin)
	app.Get("/api/v1/auth/github/callback", authHandler.GitHubCallback)

	// Protected routes
	app.Get("/api/v1/auth/me", auth.JWTMiddleware(cfg.JWTSecret), authHandler.Me)

	// File routes (protected)
	app.Post("/api/v1/files/upload", auth.JWTMiddleware(cfg.JWTSecret), fileHandler.Upload)
	app.Get("/api/v1/files", auth.JWTMiddleware(cfg.JWTSecret), fileHandler.ListFiles)
	app.Get("/api/v1/files/:id/download", auth.JWTMiddleware(cfg.JWTSecret), fileHandler.Download)
	app.Delete("/api/v1/files/:id", auth.JWTMiddleware(cfg.JWTSecret), fileHandler.Delete)
	app.Patch("/api/v1/files/share", auth.JWTMiddleware(cfg.JWTSecret), fileHandler.Share)

	// OAuth callback - Google/GitHub redirect thẳng đến đây, xử lý code và trả HTML
	app.Get("/auth/google/callback", func(c *fiber.Ctx) error {
		code := c.Query("code")
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Authenticating...</title>
    <script>
        async function completeAuth() {
            try {
                const response = await fetch('/api/v1/auth/google/callback?code=%s');
                const data = await response.json();
                if (data.token) {
                    localStorage.setItem('aeshield_token', data.token);
                    localStorage.setItem('aeshield_user', JSON.stringify(data.user));
                    window.location.href = '/dashboard';
                } else {
                    document.body.innerHTML = '<p style="color:red;text-align:center;margin-top:50px;font-family:sans-serif;">Authentication failed: ' + (data.error || 'Unknown error') + '</p><center><a href="/">Try again</a></center>';
                }
            } catch (err) {
                document.body.innerHTML = '<p style="color:red;text-align:center;margin-top:50px;font-family:sans-serif;">Error: ' + err.message + '</p><center><a href="/">Try again</a></center>';
            }
        }
        completeAuth();
    </script>
</head>
<body><p style="text-align:center;margin-top:50px;font-family:sans-serif;">Authenticating...</p></body>
</html>`, code)
		return c.Type("html").SendString(html)
	})

	app.Get("/auth/github/callback", func(c *fiber.Ctx) error {
		code := c.Query("code")
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Authenticating...</title>
    <script>
        async function completeAuth() {
            try {
                const response = await fetch('/api/v1/auth/github/callback?code=%s');
                const data = await response.json();
                if (data.token) {
                    localStorage.setItem('aeshield_token', data.token);
                    localStorage.setItem('aeshield_user', JSON.stringify(data.user));
                    window.location.href = '/dashboard';
                } else {
                    document.body.innerHTML = '<p style="color:red;text-align:center;margin-top:50px;font-family:sans-serif;">Authentication failed: ' + (data.error || 'Unknown error') + '</p><center><a href="/">Try again</a></center>';
                }
            } catch (err) {
                document.body.innerHTML = '<p style="color:red;text-align:center;margin-top:50px;font-family:sans-serif;">Error: ' + err.message + '</p><center><a href="/">Try again</a></center>';
            }
        }
        completeAuth();
    </script>
</head>
<body><p style="text-align:center;margin-top:50px;font-family:sans-serif;">Authenticating...</p></body>
</html>`, code)
		return c.Type("html").SendString(html)
	})

	// Serve frontend static files
	app.Static("/", frontendDist)

	// For SPA - redirect non-API routes to index.html
	app.Use(func(c *fiber.Ctx) error {
		path := c.Route().Path
		if !startsWith(path, "/api") && !startsWith(path, "/auth") && !startsWith(path, "/docs") {
			return c.SendFile(filepath.Join(frontendDist, "index.html"))
		}
		return c.Next()
	})

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

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
