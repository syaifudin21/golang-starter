package repository

import (
	"database/sql"
	"errors"
	"exam/internal/model"
	"fmt"
	"time"
)

// UserRepository defines methods for interacting with user data in the database.
type UserRepository interface {
	GetUserByEmail(email string) (*model.User, error)
	GetUserByID(userID int) (*model.User, error)
	GetUserByUUID(userUUID string) (*model.User, error)
	UpdateUser(userID int, name string, phone *string) error
	UpdatePassword(userID int, hashedPassword string) error
	UpdateUserRole(userID int, newRole string) error // Added UpdateUserRole method
	ListAllUsers() ([]model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow("SELECT id, uuid, name, email, phone, password, role, created_at, updated_at FROM users WHERE email = ?", email).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found, return nil user and nil error
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (r *userRepository) GetUserByID(userID int) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow("SELECT id, uuid, name, email, phone, password, role, created_at, updated_at FROM users WHERE id = ?", userID).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found, return nil user and nil error
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

func (r *userRepository) GetUserByUUID(userUUID string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow("SELECT id, uuid, name, email, phone, password, role, created_at, updated_at FROM users WHERE uuid = ?", userUUID).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found, return nil user and nil error
		}
		return nil, fmt.Errorf("failed to get user by UUID: %w", err)
	}
	return user, nil
}

func (r *userRepository) UpdateUser(userID int, name string, phone *string) error {
	_, err := r.db.Exec("UPDATE users SET name = ?, phone = ?, updated_at = ? WHERE id = ?", name, phone, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *userRepository) UpdatePassword(userID int, hashedPassword string) error {
	_, err := r.db.Exec("UPDATE users SET password = ?, updated_at = ? WHERE id = ?", hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateUserRole(userID int, newRole string) error {
	_, err := r.db.Exec("UPDATE users SET role = ?, updated_at = ? WHERE id = ?", newRole, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}
	return nil
}

func (r *userRepository) ListAllUsers() ([]model.User, error) {
	rows, err := r.db.Query("SELECT id, uuid, name, email, phone, password, role, created_at, updated_at FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.UUID,
			&user.Name,
			&user.Email,
			&user.Phone,
			&user.Password,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}