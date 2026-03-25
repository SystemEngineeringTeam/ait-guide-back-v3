# ==== 開発環境 ====
FROM golang:1.26.1:latest AS dev

WORKDIR /app

RUN go install github.com/air-verse/air@latest && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

RUN --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download

CMD ["air"]

# ==== ビルド ====
FROM golang:1.26.1:latest AS builder

WORKDIR /app

RUN --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download

RUN --mount=type=bind,target=. \
    go build -ldflags="-s" -trimpath -o /bin/server cmd/server/main.go

# ==== 本番環境 ====
FROM gcr.io/distroless/static-debian13:nonroot:latest AS production

WORKDIR /app

COPY --from=builder /bin/server .
COPY db ./db

ENV TZ=Asia/Tokyo

EXPOSE 8080

CMD ["./server"]
