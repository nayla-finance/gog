package serve

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/PROJECT_NAME/docs"
	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/registry"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		RunE:  Run,
	}

	cmd.Flags().StringP("config", "c", "config.yaml", "config file")

	return cmd
}

// @Title						PROJECT_NAME
// @Version					    1.0
// @Description				    API for PROJECT_NAME
// @BasePath					/api
// @SecurityDefinitions.apikey	ApiKey
// @In							header
// @Name						X-API-KEY
// @Description			    	API key for authentication
func Run(cmd *cobra.Command, args []string) error {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("❌ Failed to get config file: %v", err)
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("❌ Failed to load configuration: %v", err)
	}

	r := registry.NewRegistry(cfg)

	app := NewApp(cfg, r)
	app.Use(NewSwagger(cfg))

	if err := r.InitializeWithFiber(app); err != nil {
		return err
	}

	// Create error channel to capture server errors
	serverErr := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", cfg.App.Port)); err != nil {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		r.Logger().Info("Received shutdown signal", "signal", sig)

		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Start cleanup in a goroutine
		done := make(chan struct{})
		go func() {
			defer close(done)

			// Cleanup registry (your services, DB connections, etc)
			r.Cleanup()

			// Graceful shutdown of the fiber app
			if err := app.ShutdownWithContext(shutdownCtx); err != nil {
				r.Logger().Error("Error during HTTP server shutdown", "error", err)
			}
		}()

		// Wait for cleanup to finish or timeout
		select {
		case <-done:
			r.Logger().Info("Graceful shutdown completed")
		case <-shutdownCtx.Done():
			r.Logger().Error("Shutdown timed out")
		}
	}

	return nil

}

func NewApp(cfg *config.Config, r *registry.Registry) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: r.ErrorHandler().Handle,
		// Handle timeouts
		ReadTimeout:  time.Duration(cfg.App.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.App.WriteTimeout) * time.Second,

		// print all routes with their method, path and handler
		EnablePrintRoutes: true,
	})
}

func NewSwagger(cfg *config.Config) fiber.Handler {
	cacheAge := 0
	if cfg.App.Env == "development" {
		cacheAge = 0
	}

	return swagger.New(swagger.Config{
		BasePath: "/api",
		Title:    cfg.App.Name,
		Path:     "docs",
		FilePath: "./docs/swagger.json",
		CacheAge: cacheAge,
	})
}
