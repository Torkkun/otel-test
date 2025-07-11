package env

import (
	"os"
	"strings"
)

// モードごとのHandlerの設定を変える
type Mode string

const (
	Default      = "default"
	GCPOtel Mode = "gcpotel"
)

// 環境変数からモードを取得する
func GetModeFromEnv() Mode {
	env := strings.ToLower(os.Getenv("MODE"))
	switch env {
	case "otel":
		return GCPOtel
	default:
		return Default
	}
}
