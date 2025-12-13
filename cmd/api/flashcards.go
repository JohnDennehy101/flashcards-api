package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"flashcards-api.johndennehy101.tech/internal/data"
	"flashcards-api.johndennehy101.tech/internal/validator"
)

type flashcardInput struct {
	ID          int64              `json:"id"`
	Section     *string            `json:"section"`
	SectionType *string            `json:"section_type"`
	SourceFile  *string            `json:"source_file"`
	Text        string             `json:"text"`
	Question    string             `json:"question"`
	Type        data.FlashcardType `json:"flashcard_type"`
	Content     json.RawMessage    `json:"flashcard_content"`
	Categories  []string           `json:"categories"`
	Version     int32              `json:"version"`
}

func (app *application) createFlashcardHandler(w http.ResponseWriter, r *http.Request) {
	var input flashcardInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	// Decode content based on type
	var content data.FlashcardContent
	switch input.Type {
	case data.FlashcardQA:
		var qa data.QAContent
		if err := json.Unmarshal(input.Content, &qa); err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, "invalid QA content")
			return
		}
		v.Check(qa.Answer != "", "flashcard_content.answer", "answer must not be empty")
		content = qa

	case data.FlashcardMCQ:
		var mcq data.MCQContent
		if err := json.Unmarshal(input.Content, &mcq); err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, "invalid MCQ content")
			return
		}
		v.Check(len(mcq.Options) >= 2, "flashcard_content.options", "at least 2 options required")
		v.Check(mcq.CorrectIndex >= 0 && mcq.CorrectIndex < len(mcq.Options),
			"flashcard_content.correct_index", "correct index out of bounds")
		v.Check(validator.Unique(mcq.Options), "flashcard_content.options", "options must be unique")
		content = mcq

	case data.FlashcardYesNo:
		var yn data.YesNoContent
		if err := json.Unmarshal(input.Content, &yn); err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, "invalid Yes/No content")
			return
		}
		content = yn

	default:
		app.badRequestResponse(w, r, errors.New("invalid flashcard type"))
		return
	}

	flashcard := data.Flashcard{
		ID:          input.ID,
		Section:     input.Section,
		SectionType: input.SectionType,
		SourceFile:  input.SourceFile,
		Text:        input.Text,
		Question:    input.Question,
		Type:        input.Type,
		Content:     content,
		Categories:  input.Categories,
		Version:     input.Version,
		CreatedAt:   time.Now(),
	}

	if data.ValidateFlashcard(v, &flashcard); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	app.writeJSON(w, http.StatusCreated, flashcard, nil)
}

func (app *application) showFlashcardHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	sectionTitle := "Order 40F - Intellectual Property"
	sectionType := "order"
	sourceFile := "Court Rules"
	fullText := `## **1 Definitions**

1. In this Order, unless the context or subject matter otherwise requires—

the “**1992 Act**” means the [Patents Act 1992](https://www.irishstatutebook.ie/eli/1992/act/1/enacted/en/html) (No. 1 of 1992);
the “**1996 Act**” means the [Trade Marks Act 1996](https://www.irishstatutebook.ie/eli/1996/act/6/enacted/en/html) (No. 6 of 1996);
the “**2000 Act**” means the [Copyright and Related Rights Act 2000](https://www.irishstatutebook.ie/eli/2000/act/28/enacted/en/html) (No. 28 of 2000);
the “**2001 Act**” means the [Industrial Designs Act 2001](https://www.irishstatutebook.ie/eli/2001/act/39/enacted/en/html) (No. 39 of 2001);
the “**2019 Act**” means the [Copyright and Other Intellectual Property Law Provisions Act 2019](https://www.irishstatutebook.ie/eli/2019/act/19/enacted/en/html) (No. 19 of 2019);
“**intellectual property claim**” has the same meaning as in section 2 of the [2000 Act](https://www.irishstatutebook.ie/eli/2000/act/28/enacted/en/html).`

	flashcard := data.Flashcard{
		ID:          id,
		CreatedAt:   time.Now(),
		Section:     &sectionTitle,
		SectionType: &sectionType,
		SourceFile:  &sourceFile,
		Text:        fullText,

		Question: "Is the Copyright and Related Rights Act 2000 mentioned in the definitions?",
		Type:     data.FlashcardYesNo,
		Content: data.YesNoContent{
			Correct:       true,
			Justification: "the “2000 Act” means the [Copyright and Related Rights Act 2000](https://www.irishstatutebook.ie/eli/2000/act/28/enacted/en/html) (No. 28 of 2000);",
		},
		Categories: []string{"intellectual property"},
		Version:    1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"flashcard": flashcard}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
