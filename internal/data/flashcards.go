package data

import (
	"time"
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
