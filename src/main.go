package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"otel-test/env"
	"otel-test/o11y"
	"otel-test/server"
)

func main() {
	ctx := context.Background()
	// 環境変数の取得
	mode := env.GetModeFromEnv()

	shutdown, err := o11y.SetupObservability(ctx, mode)
	if err != nil {
		slog.ErrorContext(ctx, "error setting up OpenTelemetry", slog.Any("error", err))
		os.Exit(1)
	}
	// loggingとOpenTelemetryのSetUPはサーバーと切り離すのがよさそうかな
	slog.InfoContext(ctx, "server starting...")
	if err = errors.Join(server.Run(ctx, mode), shutdown(ctx)); err != nil {
		slog.ErrorContext(ctx, "server exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}
