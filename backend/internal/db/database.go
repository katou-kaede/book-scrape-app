package db

// DBへの接続管理、テーブル作成

import (
	"log"
	"time"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // ドライバは必須
)

// NewDatabase はDB接続を初期化し、sqlx.DBを返します
func NewDatabase() (*sqlx.DB, error) {
	// 環境変数からパスを取得。設定がなければデフォルト値を使う
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/books.db"
	}

	// 1. Open: 接続の準備
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// 2. Ping: 実際に接続できるか確認（現場では必須！）
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 3. 現場で必須の「コネクションプール」設定
	// SQLiteなので1でも動きますが、並列処理をするGoの現場では必ず調整します
	db.SetMaxOpenConns(25)                 // 最大接続数
	db.SetMaxIdleConns(25)                 // アイドル状態（待機中）の最大接続数
	db.SetConnMaxLifetime(5 * time.Minute) // 接続を使い回す最大時間

	// 4. テーブル作成 (マイグレーションの簡易版)
	// 現場の本格的なプロジェクトでは golang-migrate などの外部ツールを使う
	schema := `
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		price TEXT,
		stock TEXT
	);`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	log.Printf("Connected to SQLite at: %s", dbPath)
	return db, nil
}