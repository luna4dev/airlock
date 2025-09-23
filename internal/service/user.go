package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/luna4dev/airlock/internal/model"
)

func (s *SQLiteService) GetAllUsers(ctx context.Context) ([]*model.Luna4User, error) {
	log.Printf("GetAllUsers: Starting to fetch all users")
	query := `
		SELECT id, email, status, created_at, updated_at
		FROM luna4_users
		ORDER BY created_at DESC
	`

	log.Printf("GetAllUsers: Executing query: %s", query)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("GetAllUsers: Query failed with error: %v", err)
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*model.Luna4User
	log.Printf("GetAllUsers: Starting to scan rows")
	for rows.Next() {
		var user model.Luna4User

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			log.Printf("GetAllUsers: Failed to scan user row: %v", err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		log.Printf("GetAllUsers: Successfully scanned user with ID: %s", user.ID)
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetAllUsers: Error during row iteration: %v", err)
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	log.Printf("GetAllUsers: Successfully retrieved %d users", len(users))
	return users, nil
}

func (s *SQLiteService) CreateUser(ctx context.Context, user *model.Luna4User) error {
	log.Printf("CreateUser: Creating user with ID: %s, Email: %s", user.ID, user.Email)
	query := `
		INSERT INTO luna4_users (id, email, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	log.Printf("CreateUser: Executing insert query")
	_, err := s.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		log.Printf("CreateUser: Failed to create user: %v", err)
	} else {
		log.Printf("CreateUser: Successfully created user with ID: %s", user.ID)
	}
	return err
}

func (s *SQLiteService) GetUserByID(ctx context.Context, userID string) (*model.Luna4User, error) {
	log.Printf("GetUserByID: Looking for user with ID: %s", userID)
	query := `
		SELECT id, email, status, created_at, updated_at 
		FROM luna4_users
		WHERE id = ?
	`

	log.Printf("GetUserByID: Executing query")
	row := s.db.QueryRowContext(ctx, query, userID)

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
			log.Printf("GetUserByID: No user found with ID: %s", userID)
			return nil, nil
		}
		log.Printf("GetUserByID: Failed to scan user: %v", err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	log.Printf("GetUserByID: Successfully found user with ID: %s, Email: %s", user.ID, user.Email)
	return &user, nil
}

func (s *SQLiteService) GetUserByEmail(ctx context.Context, email string) (*model.Luna4User, error) {
	log.Printf("GetUserByEmail: Looking for user with email: %s", email)
	query := `
		SELECT *
		FROM luna4_users
		WHERE luna4_users.email = ?
		LIMIT 1
	`

	log.Printf("GetUserByEmail: Executing query")
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
			log.Printf("GetUserByEmail: No user found with email: %s", email)
			return nil, nil
		}
		log.Printf("GetUserByEmail: Failed to scan user: %v", err)
		return nil, fmt.Errorf("failed to get auth information: %w", err)
	}

	log.Printf("GetUserByEmail: Successfully found user with ID: %s for email: %s", user.ID, email)
	return &user, nil
}

func (s *SQLiteService) UpdateUserStatus(ctx context.Context, userID string, status model.UserStatus) error {
	log.Printf("UpdateUserStatus: Updating status for user %s to %v", userID, status)
	query := `
		UPDATE luna4_users
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now().UnixMilli()
	log.Printf("UpdateUserStatus: Executing update query")
	_, err := s.db.ExecContext(ctx, query, status, now, userID)
	if err != nil {
		log.Printf("UpdateUserStatus: Failed to update user status: %v", err)
		return fmt.Errorf("failed to update user status: %w", err)
	}

	log.Printf("UpdateUserStatus: Successfully updated status for user %s", userID)
	return nil
}

func (s *SQLiteService) DeleteUser(ctx context.Context, userID string) error {
	log.Printf("DeleteUser: Deleting user with ID: %s", userID)
	query := `DELETE FROM luna4_users WHERE id = ?`

	log.Printf("DeleteUser: Executing delete query")
	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("DeleteUser: Failed to execute delete query: %v", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Printf("DeleteUser: No user found with ID: %s", userID)
		return fmt.Errorf("no user found with ID: %s", userID)
	}

	log.Printf("DeleteUser: Successfully deleted user with ID: %s (rows affected: %d)", userID, rowsAffected)
	return nil
}

func (s *SQLiteService) CreateUserService(ctx context.Context, userService *model.Luna4UserService) error {
	log.Printf("CreateUserService: Creating service %s for user %s with permission %s", userService.Service, userService.UserID, userService.Permission)
	query := `
		INSERT INTO luna4_user_service (id, user_id, service, permission, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`

	log.Printf("CreateUserService: Executing insert query")
	_, err := s.db.ExecContext(ctx, query,
		userService.ID,
		userService.UserID,
		userService.Service,
		userService.Permission,
		userService.ExpiresAt,
	)

	if err != nil {
		log.Printf("CreateUserService: Failed to create user service: %v", err)
	} else {
		log.Printf("CreateUserService: Successfully created service %s for user %s", userService.Service, userService.UserID)
	}
	return err
}

func (s *SQLiteService) GetUserServices(ctx context.Context, userID string) ([]model.Luna4UserService, error) {
	log.Printf("GetUserServices: Fetching services for user: %s", userID)
	query := `
		SELECT id, user_id, service, permission, expires_at
		FROM luna4_user_service
		WHERE user_id = ?
		ORDER BY service
	`

	log.Printf("GetUserServices: Executing query")
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("GetUserServices: Query failed: %v", err)
		return nil, fmt.Errorf("failed to query user services: %w", err)
	}
	defer rows.Close()

	var services []model.Luna4UserService
	log.Printf("GetUserServices: Starting to scan service rows")
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
			log.Printf("GetUserServices: Failed to scan service row: %v", err)
			return nil, fmt.Errorf("failed to scan user service: %w", err)
		}

		if expiresAt.Valid {
			service.ExpiresAt = &expiresAt.Int64
		}

		log.Printf("GetUserServices: Successfully scanned service: %s for user: %s", service.Service, service.UserID)
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetUserServices: Error during row iteration: %v", err)
		return nil, fmt.Errorf("error iterating over service rows: %w", err)
	}

	log.Printf("GetUserServices: Successfully retrieved %d services for user %s", len(services), userID)
	return services, nil
}

func (s *SQLiteService) DeleteUserService(ctx context.Context, serviceID string) error {
	log.Printf("DeleteUserService: Deleting service with ID: %s", serviceID)
	query := `DELETE FROM luna4_user_service WHERE id = ?`

	log.Printf("DeleteUserService: Executing delete query")
	result, err := s.db.ExecContext(ctx, query, serviceID)
	if err != nil {
		log.Printf("DeleteUserService: Failed to execute delete query: %v", err)
		return fmt.Errorf("failed to delete user service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Printf("DeleteUserService: No service found with ID: %s", serviceID)
		return fmt.Errorf("no user service found with ID: %s", serviceID)
	}

	log.Printf("DeleteUserService: Successfully deleted service with ID: %s (rows affected: %d)", serviceID, rowsAffected)
	return nil
}
