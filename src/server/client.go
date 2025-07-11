package server

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// あんま関係ない

func callSingle(ctx context.Context) error {
	// otelhttp.Get makes an http GET request, just like net/http.Get.
	// In addition, it records a span, records metrics, and propagates context.
	res, err := otelhttp.Get(ctx, "http://localhost:8080/single")
	if err != nil {
		return err
	}

	return res.Body.Close()
}
