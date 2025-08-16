FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o dist/maniplacer cmd/main.go

FROM alpine:3.20 AS runner

WORKDIR /workspace

COPY --from=builder /app/dist/maniplacer /usr/local/bin/maniplacer

ENTRYPOINT ["/bin/sh"]