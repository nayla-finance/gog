package serve

import (
	"fmt"
	"time"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	_ "github.com/project-name/docs"
	"github.com/project-name/internal/config"
	"github.com/project-name/internal/registry"
)

// @title						Your API Name
// @version					1.0
// @description				Your API Description
// @host						localhost:3000
// @BasePath					/api
// @securityDefinitions.apikey	Bearer
// @in							header
// @name						Authorization
// @description				JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
// @security					Bearer
func Run() {
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
