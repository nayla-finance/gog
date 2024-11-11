package serve

import (
	"fmt"
	"time"

	_ "github.com/PROJECT_NAME/docs"
	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/registry"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Run:   Run,
	}
}

// @title						PROJECT_NAME
// @version					1.0
// @description				Your API Description
// @host						localhost:3000
// @BasePath					/api
// @securityDefinitions.apikey	Bearer
// @in							header
// @name						Authorization
// @description				JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
// @security					Bearer
func Run(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	r, err := registry.NewRegistry(cfg)
	if err != nil {
		panic(err)
	}

	app := NewApp(cfg, r)

	app.Use(NewSwagger(cfg))
	api := app.Group("/api")

	r.RegisterMiddlewares(app)
	r.RegisterApiRoutes(api)

	app.Listen(fmt.Sprintf(":%d", cfg.Port))

}

func NewApp(cfg *config.Config, r *registry.Registry) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ErrorHandler: r.ErrorHandler().Handle,
		// Handle timeouts
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,

		// print all routes with their method, path and handler
		EnablePrintRoutes: true,
	})
}

func NewSwagger(cfg *config.Config) fiber.Handler {
	cacheAge := 0
	if cfg.Env == "development" {
		cacheAge = 0
	}

	return swagger.New(swagger.Config{
		BasePath: "/api",
		Title:    cfg.AppName,
		Path:     "docs",
		FilePath: "./docs/swagger.json",
		CacheAge: cacheAge,
	})
}
