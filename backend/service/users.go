package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gconsus/entity"
	"gconsus/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	gitHubQueryTimeout = 10 * time.Second
)

// UserService handles user business logic
type UserService struct {
	userRepo *repository.UserRepository
	teamRepo *repository.TeamRepository
	validate *validator.Validate
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository, teamRepo *repository.TeamRepository, validate *validator.Validate) *UserService {
	return &UserService{
		userRepo: userRepo,
		teamRepo: teamRepo,
		validate: validate,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req entity.CreateUserRequest) (*entity.User, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user with this username already exists
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with username %s already exists", req.Username)
	}

	user := &entity.User{
		ID:         uuid.New(),
		Username:   req.Username,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Patronymic: req.Patronymic,
		Email:      req.Email,
		PhotoURL:   req.PhotoURL,
		Position:   req.Position,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, req entity.UpdateUserRequest) (*entity.User, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields if provided
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Patronymic != nil {
		user.Patronymic = req.Patronymic
	}
	if req.Email != nil {
		user.Email = req.Email
	}
	if req.PhotoURL != nil {
		user.PhotoURL = req.PhotoURL
	}
	if req.Position != nil {
		user.Position = req.Position
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves users with optional search and active filter
func (s *UserService) ListUsers(ctx context.Context, search string, isActive *bool, limit, offset int) ([]entity.User, error) {
	// If search term is provided, use search functionality
	if search != "" {
		users, err := s.userRepo.Search(ctx, search, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search users: %w", err)
		}
		return users, nil
	}

	// Otherwise use regular list with filters
	users, err := s.userRepo.List(ctx, isActive, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// SyncStats represents statistics from user synchronization
type SyncStats struct {
	Created  int `json:"created"`
	Updated  int `json:"updated"`
	Disabled int `json:"disabled"`
	Total    int `json:"total"`
}

// SyncUsersFromExternalAPI synchronizes users from external employee API
func (s *UserService) SyncUsersFromExternalAPI(ctx context.Context) (*SyncStats, error) {
	// Placeholder implementation - this would typically integrate with an external HR/employee API
	// For now, return empty stats to prevent compilation errors
	stats := &SyncStats{
		Created:  0,
		Updated:  0,
		Disabled: 0,
		Total:    0,
	}

	// TODO: Implement actual synchronization logic with external API
	// This would involve:
	// 1. Fetching users from external API
	// 2. Comparing with existing users
	// 3. Creating new users
	// 4. Updating existing users
	// 5. Disabling users no longer in external system

	return stats, nil
}

// GetUserTeams retrieves all teams a user belongs to
func (s *UserService) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]entity.Team, error) {
	teams, err := s.teamRepo.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user teams: %w", err)
	}
	return teams, nil
}

// UserActivity retrieves user activity from GitHub/GitLab (legacy method)
func (s *Service) UserActivity(
	ctx context.Context, login string, from time.Time, to time.Time,
) (entity.UserActivityInfo, error) {
	slog.Debug("Authenticate", "login", login)

	if err := s.validate.Var(login, "required"); err != nil {
		slog.Info("Validation failed", "UserName", login, "error", err)

		return entity.UserActivityInfo{}, ParamsError{Message: err.Error()}
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, gitHubQueryTimeout)
	defer cancel()

	query, err := s.ghClient.UserActivity(ctxTimeout, login, from, to)
	if err != nil {
		slog.Error(err.Error(), "Login", login)

		return entity.UserActivityInfo{}, err
	}

	return query.User, nil
}
