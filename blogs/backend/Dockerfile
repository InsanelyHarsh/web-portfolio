FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# TARGETOS/TARGETARCH are populated automatically by `docker buildx build
# --platform linux/amd64,linux/arm64 ...`, so this cross-compiles cleanly
# for both architectures without any extra flags.
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/webserver ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/webserver ./webserver

EXPOSE 8080
CMD ["./webserver"]
