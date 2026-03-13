# Book Scrape App

Go (Echo) と React (Vite/TypeScript) を使用した、フルスタックのWebスクレイピングアプリケーションです。  
Dockerを使用して、データベース(SQLite)やブラウザ(Playwright)を含む環境を簡単に構築できます。

## 🛠 技術スタック
### Frontend
- React 18: UI ライブラリ
- TypeScript: 静的型付けによる堅牢な開発
- Vite: 高速なビルドおよび開発サーバー
- Tailwind CSS: ユーティリティファーストなスタイリング

### Backend
- Go 1.25: 高パフォーマンスなバックエンド言語
- Echo: 軽量で拡張性の高い Web フレームワーク
- Playwright for Go: ブラウザ自動化・スクレイピング
- SQLite3: 埋め込み型データベースによる軽量なデータ管理


## 🏗 プロジェクト構成
```
book-scrape-app/
├── frontend/           # React + Vite + TypeScript (Frontend)
├── backend/            # Go + Echo (Backend API & Scraper)
├── logs/               # アプリケーションログ (Git除外)
├── data/               # SQLite データベースファイル (Git除外)
├── docker-compose.yml  # Docker構成定義
└── Dockerfile          # マルチステージビルド定義
```

## 🚀 実行方法

### 前提：
Docker Desktop がインストールされ、起動していること。

### セットアップと起動
プロジェクトのルートディレクトリで以下のコマンドを実行してください。

```bash
# ビルドと起動
docker compose up --build
```

起動後、ブラウザで http://localhost:8080 にアクセスしてください。


## 🛠 開発者向け情報

### ログの確認
ログはホスト側の `./logs/app.log` にリアルタイムで出力されます。
また、最新のログはバックアップとして `app.old.log` にローテーションされます。

### データベース
SQLiteを使用しており、データは `./data/` フォルダ内に保存されます。コンテナを削除してもデータは保持されます。

### 環境変数
.env ファイル（または `docker-compose.yml`）で以下の設定を変更可能です。

- PORT: サーバーの待機ポート（デフォルト: 8080）
- LOG_PATH: ログの出力先ディレクトリ