package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

	CorrectCount int    `json:"correct_count"`
	Status       string `json:"status"`
}
type FlashcardStats struct {
	Total      int `json:"total"`
	Mastered   int `json:"mastered"`
	InProgress int `json:"in_progress"`
	NotStarted int `json:"not_started"`
}

type Category struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
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

func (m FlashcardModel) Insert(flashcard *Flashcard, userID int64) error {
	queryCard := `
       INSERT INTO flashcards (
          section, section_type, source_file, text, question,
          flashcard_type, flashcard_content, categories, version, created_at
       ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
       RETURNING id, created_at, version`

	queryProgress := `
       INSERT INTO user_flashcards (user_id, flashcard_id, correct_count, status, last_reviewed_at)
       VALUES ($1, $2, 0, 'not_started', NOW())`

	contentJSON, err := json.Marshal(flashcard.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal flashcard content: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, queryCard,
		flashcard.Section, flashcard.SectionType, flashcard.SourceFile,
		flashcard.Text, flashcard.Question, flashcard.Type,
		contentJSON, pq.Array(flashcard.Categories), flashcard.Version, time.Now(),
	).Scan(&flashcard.ID, &flashcard.CreatedAt, &flashcard.Version)

	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, queryProgress, userID, flashcard.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m FlashcardModel) Get(id int64, userID int64) (*Flashcard, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT 
            f.id, f.section, f.section_type, f.source_file, f.text, f.question,
            f.flashcard_type, f.flashcard_content, f.categories, f.version, f.created_at,
            COALESCE(uf.correct_count, 0),
            COALESCE(uf.status, 'not_started')
        FROM flashcards f
        LEFT JOIN user_flashcards uf ON f.id = uf.flashcard_id AND uf.user_id = $2
        WHERE f.id = $1`

	var flashcard Flashcard
	var contentJSON []byte

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id, userID).Scan(
		&flashcard.ID,
		&flashcard.Section,
		&flashcard.SectionType,
		&flashcard.SourceFile,
		&flashcard.Text,
		&flashcard.Question,
		&flashcard.Type,
		&contentJSON,
		pq.Array(&flashcard.Categories),
		&flashcard.Version,
		&flashcard.CreatedAt,
		&flashcard.CorrectCount,
		&flashcard.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	switch flashcard.Type {
	case FlashcardQA:
		var qa QAContent
		if err := json.Unmarshal(contentJSON, &qa); err != nil {
			return nil, fmt.Errorf("failed to unmarshal QA content: %w", err)
		}
		flashcard.Content = qa

	case FlashcardMCQ:
		var mcq MCQContent
		if err := json.Unmarshal(contentJSON, &mcq); err != nil {
			return nil, fmt.Errorf("failed to unmarshal MCQ content: %w", err)
		}
		flashcard.Content = mcq

	case FlashcardYesNo:
		var yn YesNoContent
		if err := json.Unmarshal(contentJSON, &yn); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Yes/No content: %w", err)
		}
		flashcard.Content = yn

	default:
		return nil, fmt.Errorf("unknown flashcard type: %s", flashcard.Type)
	}

	return &flashcard, nil
}

func (m FlashcardModel) Update(flashcard *Flashcard) error {
	contentJSON, err := json.Marshal(flashcard.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal flashcard content: %w", err)
	}

	query := `
		UPDATE flashcards
		SET 
			section = $1,
			section_type = $2,
			source_file = $3,
			text = $4,
			question = $5,
			flashcard_type = $6,
			flashcard_content = $7,
			categories = $8,
			version = version + 1
		WHERE id = $9 AND version = $10
		RETURNING version
	`

	args := []any{
		flashcard.Section,
		flashcard.SectionType,
		flashcard.SourceFile,
		flashcard.Text,
		flashcard.Question,
		flashcard.Type,
		contentJSON,
		pq.Array(flashcard.Categories),
		flashcard.ID,
		flashcard.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&flashcard.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m FlashcardModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM user_flashcards WHERE flashcard_id = $1", id)
	if err != nil {
		return err
	}

	query := `DELETE FROM flashcards WHERE id = $1`
	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return tx.Commit()
}

func (m FlashcardModel) GetUserStats(userID int64) (*FlashcardStats, error) {
	query := `
        SELECT 
            COUNT(*),
            COUNT(*) FILTER (WHERE status = 'mastered'),
            COUNT(*) FILTER (WHERE status = 'in_progress'),
            COUNT(*) FILTER (WHERE status = 'not_started')
        FROM user_flashcards
        WHERE user_id = $1`

	var stats FlashcardStats
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&stats.Total,
		&stats.Mastered,
		&stats.InProgress,
		&stats.NotStarted,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (m FlashcardModel) GetAll(userID int64, section, sectionType, sourceFile string, categories []string, filters Filters) ([]*Flashcard, Metadata, error) {
	query := fmt.Sprintf(`
       SELECT 
          count(*) OVER(),
          f.id, f.section, f.section_type, f.source_file, f.text, f.question,
          f.flashcard_type, f.flashcard_content, f.categories, f.version, f.created_at,
          COALESCE(uf.correct_count, 0),
          COALESCE(uf.status, 'not_started')
       FROM flashcards f
       LEFT JOIN user_flashcards uf ON f.id = uf.flashcard_id AND uf.user_id = $1
       WHERE (to_tsvector('simple', f.section) @@ plainto_tsquery('simple', $2) OR $2 = '')
       AND (LOWER(f.section_type) = LOWER($3) OR $3 = '')
       AND (LOWER(f.source_file) = LOWER($4) OR $4 = '')
       AND (f.categories @> $5 OR $5 = '{}')
       ORDER BY %s %s, f.id ASC
       LIMIT $6 OFFSET $7`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(
		ctx,
		query,
		userID,
		section,
		sectionType,
		sourceFile,
		pq.Array(categories),
		filters.limit(),
		filters.offset(),
	)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	flashcards := []*Flashcard{}

	for rows.Next() {
		var flashcard Flashcard
		var contentJSON []byte

		err := rows.Scan(
			&totalRecords,
			&flashcard.ID,
			&flashcard.Section,
			&flashcard.SectionType,
			&flashcard.SourceFile,
			&flashcard.Text,
			&flashcard.Question,
			&flashcard.Type,
			&contentJSON,
			pq.Array(&flashcard.Categories),
			&flashcard.Version,
			&flashcard.CreatedAt,
			&flashcard.CorrectCount,
			&flashcard.Status,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		switch flashcard.Type {
		case FlashcardQA:
			var qa QAContent
			if err := json.Unmarshal(contentJSON, &qa); err != nil {
				return nil, Metadata{}, err
			}
			flashcard.Content = qa

		case FlashcardMCQ:
			var mcq MCQContent
			if err := json.Unmarshal(contentJSON, &mcq); err != nil {
				return nil, Metadata{}, err
			}
			flashcard.Content = mcq

		case FlashcardYesNo:
			var yn YesNoContent
			if err := json.Unmarshal(contentJSON, &yn); err != nil {
				return nil, Metadata{}, err
			}
			flashcard.Content = yn

		default:
			return nil, Metadata{}, fmt.Errorf("unknown flashcard type: %s", flashcard.Type)
		}

		flashcards = append(flashcards, &flashcard)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return flashcards, metadata, nil
}

func (m FlashcardModel) IncrementCorrectCount(id int64, userID int64) error {
	query := `
        INSERT INTO user_flashcards (user_id, flashcard_id, correct_count, last_reviewed_at, status)
        VALUES ($1, $2, 1, NOW(), 'in_progress')
        ON CONFLICT (user_id, flashcard_id) 
        DO UPDATE SET 
            correct_count = user_flashcards.correct_count + 1,
            last_reviewed_at = NOW(),
            status = CASE 
                WHEN user_flashcards.correct_count + 1 >= 5 THEN 'mastered'
                ELSE 'in_progress'
            END
        WHERE user_flashcards.correct_count < 5`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, id)
	return err
}

func (m FlashcardModel) ResetCorrectCount(id int64, userID int64) error {
	query := `
        INSERT INTO user_flashcards (user_id, flashcard_id, correct_count, last_reviewed_at, status)
        VALUES ($1, $2, 0, NOW(), 'not_started')
        ON CONFLICT (user_id, flashcard_id) 
        DO UPDATE SET 
            correct_count = 0,
            last_reviewed_at = NOW(),
            status = 'not_started'`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, id)
	return err
}

func (m FlashcardModel) GetAllCategories(userID int64) ([]*Category, error) {
	query := `
        SELECT category, COUNT(*)
        FROM (
            SELECT unnest(f.categories) AS category
            FROM flashcards f
            INNER JOIN user_flashcards uf ON f.id = uf.flashcard_id
            WHERE uf.user_id = $1
        ) AS expanded
        GROUP BY category
        ORDER BY category ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.Name, &c.Count); err != nil {
			return nil, err
		}
		categories = append(categories, &c)
	}

	return categories, nil
}
