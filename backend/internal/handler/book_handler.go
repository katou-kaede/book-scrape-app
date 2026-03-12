package handler

// HTTPリクエストを処理する(データを外に出す)

import (
	"net/http"
	"encoding/csv"
	"strconv"

	"book-scrape-app/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

type BookHandler struct {
	repo *repository.BookRepository
}

func NewBookHandler(repo *repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

func (h *BookHandler) GetBooks(c echo.Context) error {
	books, err := h.repo.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "データの取得に失敗しました"})
	}
	return c.JSON(http.StatusOK, books)
}

func (h *BookHandler) DownloadCSV(c echo.Context) error {
	// 1. DBからデータを全件取得
    books, err := h.repo.GetAll()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch data"})
    }

	// 2. レスポンスヘッダーの設定
    // filename=... の部分がダウンロード時のファイル名になります
    c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=books_export.csv")
    c.Response().Header().Set(echo.HeaderContentType, "text/csv; charset=utf-8")

	// 3. CSV Writerの作成
    // BOM（Byte Order Mark）を最初に入れると、Excelで開いた時の文字化けを防げます
    c.Response().Write([]byte{0xEF, 0xBB, 0xBF})
    
    writer := csv.NewWriter(c.Response().Writer)
    defer writer.Flush()

	// ヘッダー行
    writer.Write([]string{"ID", "タイトル", "価格", "在庫状況"})

    // データ行
    for _, b := range books {
        writer.Write([]string{
            strconv.Itoa(b.ID),
            b.Title,
            b.Price,
            b.Stock,
        })
    }

    return nil
}