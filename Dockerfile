# --- Stage 1: Frontend Build (React/Vite) ---
FROM node:20-slim AS frontend-builder
WORKDIR /app/frontend
# 依存関係のインストールをキャッシュするため、先にpackage.jsonだけコピー
# COPY frontend/package*.json ./
# RUN npm install --include=dev
# ソースコードをコピーしてビルド (distフォルダが生成される)
COPY frontend/dist ./dist
# RUN npm run build

# --- Stage 2: Backend Build (Go) ---
# DebianベースのLinuxイメージを使用してGoをビルド
FROM golang:1.25-bookworm AS backend-builder 
WORKDIR /app/backend
# SQLite(CGO)を使うための準備
ENV CGO_ENABLED=1
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*
# 依存関係のインストール
COPY backend/go.mod backend/go.sum ./

# 1. GitのSSLチェックをオフにする
ENV GIT_SSL_NO_VERIFY=true
# 2. Goのチェックサム検証をオフにする
ENV GONOSUMDB=*
# 3. Goに特定のドメイン（または全て）の証明書エラーを無視させる
ENV GOINSECURE=*

ENV GOPROXY=direct
RUN go mod download
# ソースコードをコピー
COPY backend/ .
# 【重要】Stage 1 で作った dist を Go のビルドディレクトリにコピー
# COPY --from=frontend-builder /app/frontend/dist ./dist
# Goをビルド
RUN go build -o /app/server ./cmd/api/main.go

# --- Stage 3: Runtime (実行用軽量イメージ) ---
FROM debian:bookworm-slim
WORKDIR /app

# SQLiteとPlaywrightに必要なライブラリをインストール
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    libnss3 libnspr4 libatk1.0-0 libatk-bridge2.0-0 libcups2 libdrm2 \
    libxkbcommon0 libxcomposite1 libxdamage1 libxfixes3 libxrandr2 \
    libgbm1 libasound2 \
    && rm -rf /var/lib/apt/lists/*

# ビルド成果物（バイナリと静的ファイル）をコピー
COPY --from=frontend-builder /app/frontend/dist ./dist
COPY --from=backend-builder /app/server .

# データとログの保存先ディレクトリを作成
RUN mkdir -p /app/data /app/logs

EXPOSE 8080

# サーバーを起動する前にブラウザをインストールする
RUN ./server --install-browsers

# サーバーを起動
CMD ["./server"]