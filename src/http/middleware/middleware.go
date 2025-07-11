package middleware

import "net/http"

// なんとなく作成した特にいらない
// Middleware結合用の関数
func ComposeMiddlewares(h http.HandlerFunc, mws ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
