package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/luna4dev/airlock/internal/model"
)

func (s *SQLiteService) GetAllUsers(ctx context.Context) ([]*model.Luna4User, error) {
	query := `
		SELECT id, email, status, created_at, updated_at, last_login_at
		FROM luna4_users
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*model.Luna4User
	for rows.Next() {
		var user model.Luna4User
		var lastLoginAt sql.NullInt64

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return users, nil
}

func (s *SQLiteService) CreateUser(ctx context.Context, user *model.Luna4User) error {
	query := `
		INSERT INTO luna4_users (id, email, status, created_at, updated_at, last_login_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (s *SQLiteService) GetUserByID(ctx context.Context, userID string) (*model.Luna4User, error) {
	query := `
		SELECT id, email, status, created_at, updated_at, last_login_at
		FROM luna4_users
		WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, userID)

	var user model.Luna4User
	var lastLoginAt sql.NullInt64

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

func (s *SQLiteService) GetUserByEmail(ctx context.Context, email string) (*model.Luna4User, error) {
	query := `
		SELECT * 
		FROM luna4_users
		WHERE luna4_users.email = ?
		LIMIT 1
	`

	row := s.db.QueryRowContext(ctx, query, email)

	var user model.Luna4User

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get auth information: %w", err)
	}

	return &user, nil
}

func (s *SQLiteService) UpdateUserStatus(ctx context.Context, userID string, status model.UserStatus) error {
	query := `
		UPDATE luna4_users
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now().UnixMilli()
	_, err := s.db.ExecContext(ctx, query, status, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

func (s *SQLiteService) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM luna4_users WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with ID: %s", userID)
	}

	return nil
}

func (s *SQLiteService) CreateUserService(ctx context.Context, userService *model.Luna4UserService) error {
	query := `
		INSERT INTO luna4_user_service (id, user_id, service, permission, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		userService.ID,
		userService.UserID,
		userService.Service,
		userService.Permission,
		userService.ExpiresAt,
	)

	return err
}

func (s *SQLiteService) GetUserServices(ctx context.Context, userID string) ([]model.Luna4UserService, error) {
	query := `
		SELECT id, user_id, service, permission, expires_at
		FROM luna4_user_service
		WHERE user_id = ?
		ORDER BY service
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user services: %w", err)
	}
	defer rows.Close()

	var services []model.Luna4UserService
	for rows.Next() {
		var service model.Luna4UserService
		var expiresAt sql.NullInt64

		err := rows.Scan(
			&service.ID,
			&service.UserID,
			&service.Service,
			&service.Permission,
			&expiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user service: %w", err)
		}

		if expiresAt.Valid {
			service.ExpiresAt = &expiresAt.Int64
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over service rows: %w", err)
	}

	return services, nil
}

func (s *SQLiteService) DeleteUserService(ctx context.Context, serviceID string) error {
	query := `DELETE FROM luna4_user_service WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete user service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user service found with ID: %s", serviceID)
	}

	return nil
}
