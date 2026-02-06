# Build stage（需从项目根目录构建，context 包含 backendapi 与 public）
FROM docker.mirrors.aliyun.com/library/golang:1.25.4-alpine AS builder

WORKDIR /build

COPY public/ ./public/
COPY backendapi/ ./
RUN sed -i 's|replace github.com/cryptoSelect/public => ../public|replace github.com/cryptoSelect/public => ./public|' go.mod
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOSUMDB=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/backend ./main/main.go

# Run stage
FROM docker.mirrors.aliyun.com/library/alpine:3.19

WORKDIR /app

COPY --from=builder /app/backend .
RUN mkdir -p config

EXPOSE 8080

CMD ["./backend"]
