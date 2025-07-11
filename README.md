# otel-test
GCPでOTELを試す

サンプル
https://github.com/GoogleCloudPlatform/golang-samples/blob/a2f0b67f5b39189525b22f8a838570dd1df5d287/opentelemetry/instrumentation/app/go.mod

# otelのインストール
```shell
go get `
"go.opentelemetry.io/otel" `
"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric" `
"go.opentelemetry.io/otel/exporters/stdout/stdouttrace" `
"go.opentelemetry.io/otel/exporters/stdout/stdoutlog" `
"go.opentelemetry.io/contrib/exporters/autoexport" `
"go.opentelemetry.io/contrib/propagators/autoprop" `
"go.opentelemetry.io/otel/sdk/log" `
"go.opentelemetry.io/otel/log/global" `
"go.opentelemetry.io/otel/propagation" `
"go.opentelemetry.io/otel/sdk/metric" `
"go.opentelemetry.io/otel/sdk/resource" `
"go.opentelemetry.io/otel/sdk/trace" `
"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp" `
"go.opentelemetry.io/contrib/bridges/otelslog"
```
gormのやつもある
 - https://github.com/go-gorm/opentelemetry