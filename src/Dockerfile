FROM golang:latest as builder
WORKDIR /workdir
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Set Environment Variable
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Build
RUN go build -o app

# Runtime Container
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /workdir/app .

EXPOSE 8080

CMD ["/app"]