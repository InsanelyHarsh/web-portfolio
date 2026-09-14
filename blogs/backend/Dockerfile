FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run on BUILDPLATFORM (native, no QEMU) before the cross-compiled build
# below, so a failing test fails the image build itself.
RUN go test ./...

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
