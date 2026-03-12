package main

import (
	"log"
	"net/http"
	"os"
	"io"

	"book-scrape-app/backend/internal/db"
	"book-scrape-app/backend/internal/handler"
	"book-scrape-app/backend/internal/repository"
	"book-scrape-app/backend/internal/scraper"
	// "book-scrape-app/backend/internal/model"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/joho/godotenv"
)
// ログの設定
func setupLogFile() (*os.File, error) {
	logFileName := "scraping.log"
	oldLogFileName := "scraping.old.log"

	// 1. ローテーション：古いログがあればバックアップに回す
	if _, err := os.Stat(logFileName); err == nil {
		_ = os.Remove(oldLogFileName) // 古いバックアップを削除
		_ = os.Rename(logFileName, oldLogFileName) // 現在のログをバックアップへ
	}

	// 2. 新規ログファイルの作成
	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	// 3. 出力先を「コンソール」と「ファイル」の両方に設定
	multi := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multi)

	return f, nil
}


func main() {
	// ログの設定
	f, err := setupLogFile()
    if err != nil {
        log.Fatalf("ログ設定に失敗: %v", err)
    }
    defer f.Close()

	// 最初に .env を読み込む
    if err := godotenv.Load(); err != nil {
        log.Println(".envファイルが見つかりません。デフォルト値を使用します。")
    }

	database, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer database.Close()

	repo := repository.NewBookRepository(database)

	// メイン処理：サーバー起動
	e := echo.New()

	// CORSの設定も環境変数から取れるようにする
    frontendURL := os.Getenv("FRONTEND_URL")
    if frontendURL == "" {
        frontendURL = "http://localhost:5173" // デフォルト値
    }

	// CORSの設定：フロントエンド（Reactなど）からのアクセスを許可する
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontendURL}, // Reactのデフォルトポート
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	h := handler.NewBookHandler(repo)
	s := scraper.NewScraper(repo)
	e.GET("/books", h.GetBooks)

	// スクレイピングを開始するURL（ボタン用）
	e.POST("/scrape", func(c echo.Context) error {
		isScanning, _, _, _ := s.GetStatus()
		if isScanning {
			return c.JSON(http.StatusConflict, map[string]string{"message": "現在スクレイピング実行中です。しばらくお待ちください。"})
		}
		// 現場の工夫：重い処理なので「別スレッド（go）」で走らせ、
		// フロントには「受け付けたよ！」と即座に返信します
		go s.Start() 
		return c.JSON(http.StatusAccepted, map[string]string{"message": "スクレイピングを開始しました"})
	})

	// スクレイピングの状態を返す
	e.GET("/scrape/status", func(c echo.Context) error {
		isScanning, lastError, current, total := s.GetStatus()
		return c.JSON(http.StatusOK, map[string]interface{}{
			"isScanning":   isScanning,
			"lastError":    lastError,
			"currentCount": current,
			"totalCount":   total,
		})
	})

	// CSVダウンロード用のエンドポイント
	e.GET("/books/download", h.DownloadCSV)

	// サーバーの起動アドレスも環境変数にする
    addr := os.Getenv("SERVER_ADDR")
    if addr == "" {
        addr = "127.0.0.1:8080"
    }
	
	log.Printf("サーバーを起動します: http://%s", addr)
	if err := e.Start(addr); err != nil {
		log.Fatal("サーバー起動失敗:", err)
	}
}