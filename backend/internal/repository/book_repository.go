package repository

// model.Book 専用の関数（メソッド）

import (
	"github.com/jmoiron/sqlx"
	"book-scrape-app/backend/internal/model"
)

type BookRepository struct {
	db *sqlx.DB
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Save(book *model.Book) error {
	query := `
		INSERT INTO books (title, price, stock) 
		VALUES (:title, :price, :stock)
		ON CONFLICT(title) DO UPDATE SET
		price=excluded.price,
		stock=excluded.stock
	`
	_, err := r.db.NamedExec(query, book)
	return err
}

func (r *BookRepository) GetAll() ([]model.Book, error) {
	var books []model.Book
	query := `SELECT * FROM books`
	err := r.db.Select(&books, query)
	return books, err
}