# syntax=docker/dockerfile:1

# Build stage — compiles the api binary (server, --migrate-up,
# --seed-admin, --reminders all live in this one binary).
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed

# Runtime — distroless non-root, no shell, no secrets in the image.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
COPY --from=build /out/seed /seed
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
