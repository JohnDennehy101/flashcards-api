-- +migrate Down
DELETE FROM user_flashcards WHERE flashcard_id IN (SELECT id FROM flashcards WHERE source_file = 'LSRA Code 2024' AND flashcard_type = 'qa');
DELETE FROM flashcards WHERE source_file = 'LSRA Code 2024' AND flashcard_type = 'qa';