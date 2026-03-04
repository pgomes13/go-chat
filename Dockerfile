FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

# Minimal runtime image — no shell, no package manager
FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
