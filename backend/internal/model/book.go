package model

type Book struct {
	ID    int    `json:"id" db:"id"`
	Title string `json:"title" db:"title"`
	Price string `json:"price" db:"price"`
	Stock string `json:"stock" db:"stock"`
}