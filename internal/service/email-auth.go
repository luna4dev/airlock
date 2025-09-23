package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/luna4dev/airlock/internal/model"
)

func (s *SQLiteService) CreateEmailAuth(ctx context.Context, emailAuth *model.Luna4EmailAuth) error {
	log.Printf("CreateEmailAuth: Creating email auth for user %s with ID: %s", emailAuth.UserID, emailAuth.ID)
	query := `
		INSERT INTO luna4_email_auth (id, user_id, token, sent_at, completed)
		VALUES (?, ?, ?, ?, ?)
	`

	log.Printf("CreateEmailAuth: Executing insert query")
	_, err := s.db.ExecContext(ctx, query,
		emailAuth.ID,
		emailAuth.UserID,
		emailAuth.Token,
		emailAuth.SentAt,
		emailAuth.Completed,
	)

	if err != nil {
		log.Printf("CreateEmailAuth: Failed to create email auth: %v", err)
	} else {
		log.Printf("CreateEmailAuth: Successfully created email auth with ID: %s", emailAuth.ID)
	}
	return err
}

func (s *SQLiteService) GetLatestEmailAuth(ctx context.Context, userID string) (*model.Luna4EmailAuth, error) {
	log.Printf("GetLatestEmailAuth: Looking for latest email auth for user: %s", userID)
	query := `
		SELECT
			*
		FROM luna4_email_auth
		WHERE user_id = ?
		ORDER BY sent_at DESC
		LIMIT 1
	`

	log.Printf("GetLatestEmailAuth: Executing query")
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
			log.Printf("GetLatestEmailAuth: No email auth found for user: %s", userID)
			return nil, nil
		}
		log.Printf("GetLatestEmailAuth: Failed to scan email auth: %v", err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	log.Printf("GetLatestEmailAuth: Successfully found email auth with ID: %s for user: %s", emailAuth.ID, userID)
	return &emailAuth, nil
}

func (s *SQLiteService) MarkEmailAuthCompleted(ctx context.Context, emailAuthID string) error {
	log.Printf("MarkEmailAuthCompleted: Marking email auth as completed for ID: %s", emailAuthID)
	query := `
		UPDATE luna4_email_auth
		SET completed = TRUE
		WHERE id = ?
	`

	log.Printf("MarkEmailAuthCompleted: Executing update query")
	result, err := s.db.ExecContext(ctx, query, emailAuthID)
	if err != nil {
		log.Printf("MarkEmailAuthCompleted: Failed to execute update query: %v", err)
		return fmt.Errorf("failed to mark email auth as completed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Printf("MarkEmailAuthCompleted: No email auth found with ID: %s", emailAuthID)
		return fmt.Errorf("no email auth found with ID: %s", emailAuthID)
	}

	log.Printf("MarkEmailAuthCompleted: Successfully marked email auth as completed for ID: %s (rows affected: %d)", emailAuthID, rowsAffected)
	return nil
}
