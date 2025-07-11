package server

import (
	"context"
	"net/http"
	"otel-test/env"
	"otel-test/http/middleware"
	"otel-test/server/service"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Server interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// HTTPServer はHTTPサーバーの実装（依存性注入対応版）
type HTTPServer struct {
	server      *http.Server
	mode        env.Mode
	userService *service.UserService // 追加: UserServiceの依存性
	tracer      trace.Tracer         // 追加: カスタムトレーサー
}

// Dependencies はサーバーが必要とする依存性をまとめた構造体
type Dependencies struct {
	UserService *service.UserService
	// 将来的に他のサービスも追加可能
	// OrderService *service.OrderService
	// PaymentService *service.PaymentService
}

// NewServer は新しいサーバーインスタンスを作成します（依存性注入対応）
func NewServer(mode env.Mode, deps *Dependencies) Server {
	return &HTTPServer{
		mode:        mode,
		userService: deps.UserService,
		tracer:      otel.Tracer("http-server"),
	}
}

func (s *HTTPServer) Start(ctx context.Context) error {
	// ハンドラーの設定
	mh := newHandler(s.mode)
	mh.handleHTTP("/single", handlerSingle())
	mh.handleHTTP("/multi", handlerMulti())

	mh.handleHTTP("/users", s.handleUsers())
	mh.handleHTTP("/users/{id}", s.handleUserByID())
	mh.handleHTTP("/health", s.handleHealth())

	// HTTPサーバーの作成
	s.server = &http.Server{
		Addr:    ":8080",
		Handler: mh.mux,
	}

	// サーバーの起動
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

type MyHandler struct {
	mux         *http.ServeMux
	wrapHandler func(http.HandlerFunc, string) http.Handler
}

func newHandler(mode env.Mode) *MyHandler {
	var wrapper func(http.HandlerFunc, string) http.Handler
	switch mode {
	case env.GCPOtel:
		wrapper = func(h http.HandlerFunc, route string) http.Handler {
			return otelhttp.NewHandler(otelhttp.WithRouteTag(route, h), route)
		}
	default:
		wrapper = func(h http.HandlerFunc, _ string) http.Handler {
			return h
		}
	}
	return &MyHandler{
		mux:         http.NewServeMux(),
		wrapHandler: wrapper,
	}
}

func (mh *MyHandler) handleHTTP(route string, handleFn http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	handler := middleware.ComposeMiddlewares(handleFn, middlewares...)
	mh.mux.Handle(route, mh.wrapHandler(handler, route))
}
