package o11y

import (
	"context"
	"otel-test/env"
)

func SetupObservability(ctx context.Context, mode env.Mode) (func(context.Context) error, error) {
	setupLogging()
	switch mode {
	case env.GCPOtel:
		shutdown, err := setupOpenTelemetry(ctx)
		if err != nil {
			return nil, err
		}
		return shutdown, nil
	default:
		// no-op 関数を返す // nill errorをしないように何もしない空だけ返す
		return func(context.Context) error { return nil }, nil
	}
}
