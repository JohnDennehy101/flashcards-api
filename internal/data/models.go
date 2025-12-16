package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Flashcards FlashcardModel
	Users      UserModel
	Tokens     TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Flashcards: FlashcardModel{DB: db},
		Tokens:     TokenModel{DB: db},
		Users:      UserModel{DB: db},
	}
}
