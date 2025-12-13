package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Flashcards FlashcardModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Flashcards: FlashcardModel{DB: db},
	}
}
