package repository

import (
	"context"
	"database/sql"
	"fmt"

	"gconsus/database"
	"gconsus/entity"

	"github.com/google/uuid"
)

// TeamRepository handles team database operations
type TeamRepository struct {
	db *database.DB
}

// NewTeamRepository creates a new TeamRepository
func NewTeamRepository(db *database.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create creates a new team
func (r *TeamRepository) Create(ctx context.Context, team *entity.Team) error {
	query := `
		INSERT INTO teams (id, name, description, manager_id, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	if team.ID == uuid.Nil {
		team.ID = uuid.New()
	}

	err := r.db.QueryRowxContext(ctx, query,
		team.ID, team.Name, team.Description, team.ManagerID, team.IsActive,
	).Scan(&team.CreatedAt, &team.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

// GetByID retrieves a team by ID
func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Team, error) {
	var team entity.Team
	query := `SELECT * FROM teams WHERE id = $1`

	err := r.db.GetContext(ctx, &team, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("team not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &team, nil
}

// GetByName retrieves a team by name
func (r *TeamRepository) GetByName(ctx context.Context, name string) (*entity.Team, error) {
	var team entity.Team
	query := `SELECT * FROM teams WHERE name = $1`

	err := r.db.GetContext(ctx, &team, query, name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("team not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &team, nil
}

// Update updates a team
func (r *TeamRepository) Update(ctx context.Context, team *entity.Team) error {
	query := `
		UPDATE teams
		SET name = $1, description = $2, manager_id = $3, is_active = $4
		WHERE id = $5
	`

	result, err := r.db.ExecContext(ctx, query,
		team.Name, team.Description, team.ManagerID, team.IsActive, team.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("team not found")
	}

	return nil
}

// List retrieves all teams
func (r *TeamRepository) List(ctx context.Context, isActive *bool, limit, offset int) ([]entity.Team, error) {
	query := `SELECT * FROM teams`
	args := []interface{}{}
	argCount := 0

	if isActive != nil {
		argCount++
		query += fmt.Sprintf(" WHERE is_active = $%d", argCount)
		args = append(args, *isActive)
	}

	query += " ORDER BY name"

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

	var teams []entity.Team
	err := r.db.SelectContext(ctx, &teams, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	return teams, nil
}

// Delete soft deletes a team
func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE teams SET is_active = false WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("team not found")
	}

	return nil
}

// AddMember adds a member to a team
func (r *TeamRepository) AddMember(ctx context.Context, member *entity.TeamMember) error {
	query := `
		INSERT INTO team_members (id, team_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	if member.ID == uuid.Nil {
		member.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		member.ID, member.TeamID, member.UserID, member.Role, member.JoinedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add team member: %w", err)
	}

	return nil
}

// RemoveMember removes a member from a team (soft delete)
func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	query := `
		UPDATE team_members
		SET left_at = NOW()
		WHERE team_id = $1 AND user_id = $2 AND left_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("team member not found or already removed")
	}

	return nil
}

// GetMembers retrieves all active members of a team
func (r *TeamRepository) GetMembers(ctx context.Context, teamID uuid.UUID) ([]entity.TeamMember, error) {
	query := `
		SELECT tm.*, u.username, u.first_name, u.last_name, u.patronymic,
		       u.email, u.photo_url, u.position, u.is_active,
		       u.created_at as "user.created_at", u.updated_at as "user.updated_at"
		FROM team_members tm
		JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1 AND tm.left_at IS NULL
		ORDER BY tm.joined_at
	`

	rows, err := r.db.QueryxContext(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	var members []entity.TeamMember
	for rows.Next() {
		var member entity.TeamMember
		var user entity.User

		err := rows.Scan(
			&member.ID, &member.TeamID, &member.UserID, &member.Role,
			&member.JoinedAt, &member.LeftAt,
			&user.Username, &user.FirstName, &user.LastName, &user.Patronymic,
			&user.Email, &user.PhotoURL, &user.Position, &user.IsActive,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}

		user.ID = member.UserID
		member.User = &user
		members = append(members, member)
	}

	return members, nil
}

// GetTeamWithMembers retrieves a team with all its members
func (r *TeamRepository) GetTeamWithMembers(ctx context.Context, teamID uuid.UUID) (*entity.TeamWithMembers, error) {
	team, err := r.GetByID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	members, err := r.GetMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}

	teamWithMembers := &entity.TeamWithMembers{
		Team:    *team,
		Members: members,
	}

	// Get manager info if exists
	if team.ManagerID != nil {
		userRepo := NewUserRepository(r.db)
		manager, err := userRepo.GetByID(ctx, *team.ManagerID)
		if err == nil {
			teamWithMembers.Manager = manager
		}
	}

	return teamWithMembers, nil
}

// GetUserTeams retrieves all teams a user belongs to
func (r *TeamRepository) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]entity.Team, error) {
	query := `
		SELECT t.* FROM teams t
		JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1 AND tm.left_at IS NULL AND t.is_active = true
		ORDER BY t.name
	`

	var teams []entity.Team
	err := r.db.SelectContext(ctx, &teams, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user teams: %w", err)
	}

	return teams, nil
}

// IsMember checks if a user is a member of a team
func (r *TeamRepository) IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM team_members
			WHERE team_id = $1 AND user_id = $2 AND left_at IS NULL
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, teamID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check team membership: %w", err)
	}

	return exists, nil
}

// UpdateMemberRole updates a team member's role
func (r *TeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	query := `
		UPDATE team_members
		SET role = $1
		WHERE team_id = $2 AND user_id = $3 AND left_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, role, teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("team member not found")
	}

	return nil
}
