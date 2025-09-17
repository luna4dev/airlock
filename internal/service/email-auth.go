package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/luna4dev/airlock/internal/model"
)

func (s *SQLiteService) CreateEmailAuth(ctx context.Context, emailAuth *model.Luna4EmailAuth) error {

	query := `
		INSERT INTO luna4_email_auth (id, user_id, token, sent_at, completed)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		emailAuth.ID,
		emailAuth.UserID,
		emailAuth.Token,
		emailAuth.SentAt,
		emailAuth.Completed,
	)

	return err
}

func (s *SQLiteService) GetLatestEmailAuth(ctx context.Context, userID string) (*model.Luna4EmailAuth, error) {
	query := `
		SELECT 
			* 
		FROM luna4_email_auth
		WHERE user_id = ?
		ORDER BY sent_at DESC
		LIMIT 1
	`

	row := s.db.QueryRowContext(ctx, query, userID)

	var emailAuth model.Luna4EmailAuth

	err := row.Scan(
		&emailAuth.ID,
		&emailAuth.UserID,
		&emailAuth.Token,
		&emailAuth.SentAt,
		&emailAuth.Completed,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &emailAuth, nil
}

func (s *SQLiteService) MarkEmailAuthCompleted(ctx context.Context, emailAuthID string) error {
	query := `
		UPDATE luna4_email_auth
		SET completed = TRUE
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query, emailAuthID)
	if err != nil {
		return fmt.Errorf("failed to mark email auth as completed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no email auth found with ID: %s", emailAuthID)
	}

	return nil
}
