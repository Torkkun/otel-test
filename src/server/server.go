package server

import (
	"context"
	"net/http"
	"otel-test/env"
	"otel-test/http/middleware"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Run(ctx context.Context, mode env.Mode) error {

	// set up handler
	// モードによってハンドラーの設定を変える
	mh := NewHandler(mode)
	mh.handleHTTP("/single", handlerSingle())
	mh.handleHTTP("/multi", handlerMulti())

	return http.ListenAndServe(":8080", mh.mux)
}

type MyHandler struct {
	mux         *http.ServeMux
	wrapHandler func(http.HandlerFunc, string) http.Handler
}

func NewHandler(mode env.Mode) *MyHandler {
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
