package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"flashcards-api.johndennehy101.tech/internal/validator"
	"github.com/lib/pq"
)

type FlashcardType string

const (
	FlashcardQA    FlashcardType = "qa"
	FlashcardMCQ   FlashcardType = "mcq"
	FlashcardYesNo FlashcardType = "yes_no"
)

type FlashcardContent interface {
	isFlashcardContent()
}

type YesNoContent struct {
	Correct       bool   `json:"correct"`
	Justification string `json:"justification,omitempty"`
}

func (YesNoContent) isFlashcardContent() {}

type QAContent struct {
	Answer        string `json:"answer"`
	Justification string `json:"justification,omitempty"`
}

func (QAContent) isFlashcardContent() {}

type MCQContent struct {
	Options       []string `json:"options"`
	CorrectIndex  int      `json:"correct_index"`
	Justification string   `json:"justification,omitempty"`
}

func (MCQContent) isFlashcardContent() {}

type Flashcard struct {
	ID int64 `json:"id"`

	// specific name or identifier (“Chapter 3”, “HC Order 2023/114”)
	Section *string `json:"section"`

	// “chapter” / “court_order”
	SectionType *string `json:"section_type"`

	// e.g., "Foundation Manual", "Court Rules"
	SourceFile *string `json:"source_file"`

	Text string `json:"text"`

	CreatedAt time.Time `json:"-"`

	Question string           `json:"question"`
	Type     FlashcardType    `json:"flashcard_type"`
	Content  FlashcardContent `json:"flashcard_content"`

	Categories []string `json:"categories"`
	Version    int32    `json:"version"`
}

func ValidateFlashcard(v *validator.Validator, flashcard *Flashcard) {
	v.Check(flashcard.Question != "", "question", "question must be provided")
	v.Check(flashcard.Text != "", "text", "text must be provided")
	v.Check(validator.Unique(flashcard.Categories), "categories", "categories must be unique")
	v.Check(validator.PermittedValue(flashcard.Type, FlashcardQA, FlashcardMCQ, FlashcardYesNo),
		"flashcard_type", "invalid flashcard type")
}

type FlashcardModel struct {
	DB *sql.DB
}

func (m FlashcardModel) Insert(flashcard *Flashcard) error {
	query := `
		INSERT INTO flashcards (
			section, section_type, source_file, text, question,
			flashcard_type, flashcard_content, categories, version, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, version`
	
	contentJSON, err := json.Marshal(flashcard.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal flashcard content: %w", err)
	}

	args := []any{
		flashcard.Section,
		flashcard.SectionType,
		flashcard.SourceFile,
		flashcard.Text,
		flashcard.Question,
		flashcard.Type,
		contentJSON,
		pq.Array(flashcard.Categories),
		flashcard.Version,
		time.Now(),
	}

	return m.DB.QueryRow(query, args...).Scan(&flashcard.ID, &flashcard.CreatedAt, &flashcard.Version)
}

func (m FlashcardModel) Get(id int64) (*Flashcard, error) {
	return nil, nil
}

func (m FlashcardModel) Update(flashcard *Flashcard) error {
	return nil
}

func (m FlashcardModel) Delete(id int64) error {
	return nil
}
