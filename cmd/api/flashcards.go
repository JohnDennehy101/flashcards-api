package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

	user := app.contextGetUser(r)

	err = app.models.Flashcards.Insert(&flashcard, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/flashcards/%d", flashcard.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"flashcard": flashcard}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showFlashcardHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)

	flashcard, err := app.models.Flashcards.Get(id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"flashcard": flashcard}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showFlashcardStatsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	stats, err := app.models.Flashcards.GetUserStats(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"stats": stats}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateFlashcardHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)

	flashcard, err := app.models.Flashcards.Get(id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
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

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var content data.FlashcardContent
	switch input.Type {
	case data.FlashcardQA:
		var qa data.QAContent
		if err := json.Unmarshal(input.Content, &qa); err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, "invalid QA content")
			return
		}
		content = qa

	case data.FlashcardMCQ:
		var mcq data.MCQContent
		if err := json.Unmarshal(input.Content, &mcq); err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, "invalid MCQ content")
			return
		}
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

	flashcard.Section = input.Section
	flashcard.SectionType = input.SectionType
	flashcard.SourceFile = input.SourceFile
	flashcard.Text = input.Text
	flashcard.Question = input.Question
	flashcard.Type = input.Type
	flashcard.Content = content
	flashcard.Categories = input.Categories

	v := validator.New()

	if data.ValidateFlashcard(v, flashcard); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Flashcards.Update(flashcard)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"flashcard": flashcard}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listFlashcardsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	qs := r.URL.Query()
	v := validator.New()

	categories := app.readCSV(qs, "categories", []string{})
	hideMastered := app.readBool(qs, "hide_mastered", false, v)
	file := app.readString(qs, "file", "")
	section := app.readString(qs, "section", "")
	qType := app.readString(qs, "flashcard_type", "")

	paging := data.Filters{
		Page:         app.readInt(qs, "page", 1, v),
		PageSize:     app.readInt(qs, "page_size", 20, v),
		Sort:         app.readString(qs, "sort", "id"),
		SortSafelist: []string{"id", "section", "file", "-id", "-section", "-file", "random"},
	}

	if data.ValidateFilters(v, paging); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	flashcards, metadata, err := app.models.Flashcards.GetAll(
		user.ID, section, qType, file, categories, hideMastered, paging,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	filterOptions, err := app.models.Flashcards.GetFilterMetadata(user.ID, file, qType, hideMastered)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{
		"flashcards":     flashcards,
		"metadata":       metadata,
		"filter_options": filterOptions,
	}, nil)
}

func (app *application) deleteFlashcardHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Flashcards.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "flashcard successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) reviewFlashcardHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)

	err = app.models.Flashcards.IncrementCorrectCount(id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "progress updated"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) resetFlashcardHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)

	err = app.models.Flashcards.ResetCorrectCount(id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "progress reset"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
