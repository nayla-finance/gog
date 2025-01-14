package main

import (
	"fmt"
	"os"
	"time"

	"github.com/PROJECT_NAME/cmd/migrate"
	"github.com/PROJECT_NAME/cmd/serve"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
)

func main() {
	dns := os.Getenv("SENTRY__DSN")

	if dns == "" {
		fmt.Println("⚠️ SENTRY__DSN is not set in env")
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn: dns,
	})
	if err != nil {
		fmt.Println("⚠️ Failed to initialize sentry: ", err)
	}
	defer sentry.Flush(2 * time.Second)

	cmd := &cobra.Command{
		Use:   "PROJECT_NAME",
		Short: "PROJECT_NAME CLI",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	cmd.AddCommand(serve.NewServeCmd(), migrate.NewMigrateCmd())
	if err := cmd.Execute(); err != nil {
		sentry.CaptureException(err)
		panic(err)
	}
}
