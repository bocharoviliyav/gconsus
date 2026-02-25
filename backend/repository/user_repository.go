package repository

import (
	"context"
	"database/sql"
	"fmt"

	"gconsus/database"
	"gconsus/entity"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *database.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, username, first_name, last_name, patronymic, email, photo_url, position, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	err := r.db.QueryRowxContext(ctx, query,
		user.ID, user.Username, user.FirstName, user.LastName,
		user.Patronymic, user.Email, user.PhotoURL, user.Position, user.IsActive,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	query := `SELECT * FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	query := `SELECT * FROM users WHERE username = $1`

	err := r.db.GetContext(ctx, &user, query, username)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, patronymic = $3, email = $4,
		    photo_url = $5, position = $6, is_active = $7
		WHERE id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		user.FirstName, user.LastName, user.Patronymic, user.Email,
		user.PhotoURL, user.Position, user.IsActive, user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Upsert creates or updates a user based on username
func (r *UserRepository) Upsert(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, username, first_name, last_name, patronymic, email, photo_url, position, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (username) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			patronymic = EXCLUDED.patronymic,
			email = EXCLUDED.email,
			photo_url = EXCLUDED.photo_url,
			position = EXCLUDED.position,
			is_active = EXCLUDED.is_active
		RETURNING id, created_at, updated_at
	`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	err := r.db.QueryRowxContext(ctx, query,
		user.ID, user.Username, user.FirstName, user.LastName,
		user.Patronymic, user.Email, user.PhotoURL, user.Position, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	return nil
}

// List retrieves users with optional filters
func (r *UserRepository) List(ctx context.Context, isActive *bool, limit, offset int) ([]entity.User, error) {
	query := `SELECT * FROM users`
	args := []interface{}{}
	argCount := 0

	if isActive != nil {
		argCount++
		query += fmt.Sprintf(" WHERE is_active = $%d", argCount)
		args = append(args, *isActive)
	}

	query += " ORDER BY last_name, first_name"

	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	var users []entity.User
	err := r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context, isActive *bool) (int, error) {
	query := `SELECT COUNT(*) FROM users`
	args := []interface{}{}

	if isActive != nil {
		query += " WHERE is_active = $1"
		args = append(args, *isActive)
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// Delete soft deletes a user by setting is_active to false
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET is_active = false WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetByIDs retrieves multiple users by their IDs
func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.User, error) {
	if len(ids) == 0 {
		return []entity.User{}, nil
	}

	query := `SELECT * FROM users WHERE id = ANY($1)`

	var users []entity.User
	pqIDs := make([]string, len(ids))
	for i, id := range ids {
		pqIDs[i] = id.String()
	}
	err := r.db.SelectContext(ctx, &users, query, pq.Array(pqIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}

	return users, nil
}

// Search searches users by name or email
func (r *UserRepository) Search(ctx context.Context, searchTerm string, limit int) ([]entity.User, error) {
	query := `
		SELECT * FROM users
		WHERE is_active = true
		  AND (
		    first_name ILIKE $1 OR
		    last_name ILIKE $1 OR
		    email ILIKE $1 OR
		    username ILIKE $1
		  )
		ORDER BY last_name, first_name
		LIMIT $2
	`

	searchPattern := "%" + searchTerm + "%"
	var users []entity.User
	err := r.db.SelectContext(ctx, &users, query, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}
