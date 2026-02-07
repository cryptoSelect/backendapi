# Build stage
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Copy go mod files first for better cache
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOSUMDB=off
RUN go mod download

# Copy source
COPY . .

# Build binary (main entry is main/main.go)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/backend ./main/main.go

# Run stage
FROM alpine:3.19

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/backend .
RUN mkdir -p config

EXPOSE 8080

CMD ["./backend"]
