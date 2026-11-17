# 构建阶段
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /shortlink ./cmd/server

# 运行阶段
FROM alpine:3.19
RUN adduser -D appuser
COPY --from=build /shortlink /usr/local/bin/shortlink
USER appuser
EXPOSE 8080
ENTRYPOINT ["shortlink"]
