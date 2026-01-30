-- +migrate Down
DELETE FROM user_flashcards WHERE flashcard_id IN (SELECT id FROM flashcards WHERE source_file = 'Land & Conveyancing Manual' AND flashcard_type = 'mcq' AND categories @> ARRAY['AI']);
DELETE FROM flashcards WHERE source_file = 'Land & Conveyancing Manual' AND flashcard_type = 'mcq' AND categories @> ARRAY['AI'];