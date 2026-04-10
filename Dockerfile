# syntax=docker/dockerfile:1

# 第一阶段：构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# 第二阶段：构建后端
FROM golang:1.25 AS backend-builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
RUN apt-get update && apt-get install -y gcc musl-dev curl && rm -rf /var/lib/apt/lists/*
COPY go.mod ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN curl -sL "https://github.com/adobe-fonts/source-han-sans/raw/release/OTF/SimplifiedChinese/SourceHanSansSC-Regular.otf" -o /tmp/SourceHanSansSC-Regular.otf && \
    mkdir -p /app/fonts && mv /tmp/SourceHanSansSC-Regular.otf /app/fonts/
RUN go mod tidy && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bibibibi ./cmd/bibibibi

# 第三阶段：运行
FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*
COPY --from=backend-builder /app/bibibibi .
COPY --from=backend-builder /app/web/dist ./dist
COPY --from=backend-builder /app/fonts ./fonts
RUN mkdir -p /data
EXPOSE 8080
ENV BIBIBIBI_DB_PATH=/data/bibibibi.db
ENV BIBIBIBI_PORT=8080
ENV FRONTEND_DIST=/app/dist
ENV BIBIBIBI_FONT_PATH=/app/fonts/SourceHanSansSC-Regular.otf
CMD ["./bibibibi"]
