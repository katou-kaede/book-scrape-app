package handler

// HTTPリクエストを処理する(データを外に出す)

import (
	"net/http"

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