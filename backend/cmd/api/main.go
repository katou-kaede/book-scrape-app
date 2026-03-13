package main

import (
	"log"
	"net/http"
	"os"
	"flag"

	"book-scrape-app/backend/internal/db"
	"book-scrape-app/backend/internal/handler"
	"book-scrape-app/backend/internal/repository"
	"book-scrape-app/backend/internal/scraper"
	"book-scrape-app/backend/internal/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
)


func main() {
	installFlag := flag.Bool("install-browsers", false, "Install Playwright browsers")
	flag.Parse()

	if *installFlag {
		// インストール処理だけ実行
		os.Setenv("PLAYWRIGHT_BROWSER_TO_INSTALL", "chromium")

		err := playwright.Install()
		if err != nil {
			log.Fatalf("Failed to install: %v", err)
		}
		return // ★ここで return してプログラムを終了させるのが超重要！
	}

	// ログの設定
	stopLogger := logger.SetupLogFile()
	defer stopLogger()

	// 最初に .env を読み込む
    if err := godotenv.Load(); err != nil {
        log.Println(".envファイルが見つかりません。デフォルト値を使用します。")
    }

	// メイン処理：サーバー起動
	e := echo.New()

	database, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer database.Close()

	repo := repository.NewBookRepository(database)


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

	// 静的ファイルの配信
	e.Static("/", "dist")
    e.File("/", "dist/index.html")

	// サーバーの起動アドレスも環境変数にする
    addr := os.Getenv("SERVER_ADDR")
    if addr == "" {
        addr = "0.0.0.0:8080"
    }
	
	log.Printf("サーバーを起動します: http://%s", addr)
	if err := e.Start(addr); err != nil {
		log.Fatal("サーバー起動失敗:", err)
	}
}