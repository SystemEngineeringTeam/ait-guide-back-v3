# ==== 開発環境 ====
FROM golang:1.26-alpine AS dev

WORKDIR /app

RUN apk add --no-cache git

RUN go install github.com/air-verse/air@latest && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

RUN --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download

CMD ["air"]

# ==== ビルド ====
FROM golang:1.26-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0 \
    GOOS=linux

RUN --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download

RUN --mount=type=bind,target=. \
    go build -ldflags="-s -w" -trimpath -o /bin/server cmd/server/main.go

# ==== 本番環境 ====
FROM gcr.io/distroless/static-debian13:nonroot AS production

WORKDIR /app

COPY --from=builder /bin/server .
COPY db ./db

ENV TZ=Asia/Tokyo

EXPOSE 8080

CMD ["./server"]
