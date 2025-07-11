package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"otel-test/http/response"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
)

func handlerSingle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sleepTime := randomSleep(r)
		fmt.Fprintf(w, "work completed in %v\n", sleepTime)
	}
}

func handlerMulti() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subRequests := 3 + rand.Intn(4)
		// Write a structured log with the request context, which allows the log to
		// be linked with the trace for this request.
		slog.InfoContext(r.Context(), "handle /multi request", slog.Int("subRequests", subRequests))

		err := computeSubrequests(r, subRequests)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		response.Success(w, map[string]string{"multi": "success"})
	}
}

func (s *HTTPServer) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := s.tracer.Start(r.Context(), "health-check")
		defer span.End()

		span.SetAttributes(attribute.String("endpoint", "health"))

		if err := s.performHealthCheck(ctx); err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.Bool("health.check.failed", true))
			response.InternalServerError(w, "Health check failed")
			return
		}

		healthStatus := map[string]string{
			"status":  "healthy",
			"version": "1.0.0",
		}

		span.SetAttributes(attribute.Bool("health.check.success", true))
		response.Success(w, healthStatus)
	}
}

// performHealthCheck は実際のヘルスチェック処理を行います
func (s *HTTPServer) performHealthCheck(ctx context.Context) error {
	// データベース接続チェック（UserServiceが利用可能な場合）
	if s.userService != nil {
		// 簡単なクエリでデータベース接続を確認
		_, err := s.userService.ListUsers(ctx, 1, 0)
		if err != nil {
			return fmt.Errorf("database health check failed: %w", err)
		}
	}

	return nil
}

// handleUsers はユーザー一覧取得/作成エンドポイント
func (s *HTTPServer) handleUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		switch r.Method {
		case http.MethodGet:
			s.getUsersList(ctx, w, r)
		case http.MethodPost:
			s.createUser(ctx, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// getUsersList はユーザー一覧を取得
func (s *HTTPServer) getUsersList(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(ctx, "get-users-list")
	defer span.End()

	// クエリパラメータの解析
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // デフォルト値
	offset := 0 // デフォルト値

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	span.SetAttributes(
		attribute.Int("query.limit", limit),
		attribute.Int("query.offset", offset),
	)

	// サービス層の呼び出し
	users, err := s.userService.ListUsers(ctx, limit, offset)
	if err != nil {
		span.RecordError(err)
		response.InternalServerError(w, "Failed to get users")
		return
	}

	span.SetAttributes(attribute.Int("result.count", len(users)))
	response.Success(w, map[string]interface{}{
		"users":  users,
		"count":  len(users),
		"limit":  limit,
		"offset": offset,
	})
}

// createUser は新しいユーザーを作成
func (s *HTTPServer) createUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(ctx, "create-user")
	defer span.End()

	// リクエストボディの解析
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// バリデーション
	if req.Name == "" || req.Email == "" {
		span.SetAttributes(attribute.Bool("validation.failed", true))
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("user.name", req.Name),
		attribute.String("user.email", req.Email),
	)

	// サービス層の呼び出し
	user, err := s.userService.CreateUser(ctx, req.Name, req.Email)
	if err != nil {
		span.RecordError(err)
		response.InternalServerError(w, "Failed to create user")
		return
	}

	span.SetAttributes(attribute.Int("user.created_id", int(user.ID)))

	w.WriteHeader(http.StatusCreated)
	response.Success(w, user)
}

// handleUserByID は特定ユーザーの取得エンドポイント
func (s *HTTPServer) handleUserByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		switch r.Method {
		case http.MethodGet:
			s.getUserByID(ctx, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// getUserByID は特定のユーザーを取得
func (s *HTTPServer) getUserByID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(ctx, "get-user-by-id")
	defer span.End()

	// URLパスからIDを抽出（簡易実装）
	// 実際のプロダクトではgorilla/muxやgin等のルーターを使用することを推奨
	path := r.URL.Path
	idStr := path[len("/users/"):]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int("user.id", int(id)))

	// サービス層の呼び出し
	user, err := s.userService.GetUserByID(ctx, uint(id))
	if err != nil {
		span.RecordError(err)
		response.InternalServerError(w, "Failed to get user")
		return
	}

	if user == nil {
		span.SetAttributes(attribute.Bool("user.not_found", true))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response.Success(w, user)
}
