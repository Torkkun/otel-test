package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"otel-test/database"
	"otel-test/env"
	"otel-test/o11y"
	"otel-test/server"
	"otel-test/server/entity"
	"otel-test/server/repository"
	"otel-test/server/service"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	// 環境変数の取得
	mode := env.GetModeFromEnv()

	// Observabilityのセットアップ (Opentelemetry)
	otelShutdown, err := o11y.SetupObservability(ctx, mode)
	if err != nil {
		slog.ErrorContext(ctx, "error setting up OpenTelemetry", slog.Any("error", err))
		os.Exit(1)
	}

	// データベース接続
	db, err := database.NewCloudSQLDB()
	if err != nil {
		slog.ErrorContext(ctx, "failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// マイグレーション
	if err := db.AutoMigrate(&entity.User{}); err != nil {
		slog.ErrorContext(ctx, "failed to migrate database", slog.Any("error", err))
		os.Exit(1)
	}

	// リポジトリとサービスの初期化
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	// サーバー依存性の準備
	deps := &server.Dependencies{
		UserService: userService,
	}

	// サーバーの作成
	httpServer := server.NewServer(mode, deps)

	slog.InfoContext(ctx, "server starting...")

	// サーバーの起動とGraceful Shutdown
	if err := runWithGracefulShutdown(ctx, httpServer, otelShutdown); err != nil {
		slog.ErrorContext(ctx, "server exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func runWithGracefulShutdown(ctx context.Context, httpServer server.Server, otelShutdown func(context.Context) error) error {
	// シグナルを受信するためのチャネル
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// サーバーの起動
	serverErrChan := make(chan error, 1)
	go func() {
		serverErrChan <- httpServer.Start(ctx)
	}()

	// シグナルまたはサーバーエラーを待機
	select {
	case err := <-serverErrChan:
		if err != nil {
			slog.ErrorContext(ctx, "server error", slog.Any("error", err))
			return err
		}
	case sig := <-sigChan:
		slog.InfoContext(ctx, "received signal, starting graceful shutdown", slog.String("signal", sig.String()))
	}

	// Graceful Shutdownの実行
	return performGracefulShutdown(ctx, httpServer, otelShutdown)
}

func performGracefulShutdown(ctx context.Context, httpServer server.Server, otelShutdown func(context.Context) error) error {
	// シャットダウンのタイムアウト設定
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var shutdownErrors []error

	// HTTPサーバーのシャットダウン
	slog.InfoContext(shutdownCtx, "shutting down HTTP server...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "HTTP server shutdown error", slog.Any("error", err))
		shutdownErrors = append(shutdownErrors, err)
	}

	// OpenTelemetryのシャットダウン
	slog.InfoContext(shutdownCtx, "shutting down OpenTelemetry...")
	if err := otelShutdown(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "OpenTelemetry shutdown error", slog.Any("error", err))
		shutdownErrors = append(shutdownErrors, err)
	}

	if len(shutdownErrors) > 0 {
		return errors.Join(shutdownErrors...)
	}

	slog.InfoContext(shutdownCtx, "graceful shutdown completed")
	return nil
}
