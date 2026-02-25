package service

import (
	"context"
	"fmt"
	"time"

	"gconsus/entity"
	"gconsus/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// TeamService handles team business logic
type TeamService struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
	validate *validator.Validate
}

// NewTeamService creates a new team service
func NewTeamService(
	teamRepo *repository.TeamRepository,
	userRepo *repository.UserRepository,
	validate *validator.Validate,
) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		validate: validate,
	}
}

// CreateTeam creates a new team
func (s *TeamService) CreateTeam(ctx context.Context, req entity.CreateTeamRequest) (*entity.Team, error) {
	// Validate request
	if err := s.validate.Struct(req); err != nil {
		return nil, ParamsError{Message: fmt.Sprintf("validation failed: %v", err)}
	}

	// Check if team with same name already exists
	existingTeam, err := s.teamRepo.GetByName(ctx, req.Name)
	if err == nil && existingTeam != nil {
		return nil, ParamsError{Message: "team with this name already exists"}
	}

	// If manager is specified, verify they exist
	if req.ManagerID != nil {
		manager, err := s.userRepo.GetByID(ctx, *req.ManagerID)
		if err != nil || manager == nil {
			return nil, ParamsError{Message: "manager not found"}
		}
		if !manager.IsActive {
			return nil, ParamsError{Message: "manager is not active"}
		}
	}

	// Create team
	team := &entity.Team{
		Name:        req.Name,
		Description: req.Description,
		ManagerID:   req.ManagerID,
		IsActive:    true,
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return team, nil
}

// GetTeam retrieves a team by ID
func (s *TeamService) GetTeam(ctx context.Context, id uuid.UUID) (*entity.TeamWithMembers, error) {
	team, err := s.teamRepo.GetTeamWithMembers(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return team, nil
}

// ListTeams retrieves all teams with optional filtering
func (s *TeamService) ListTeams(ctx context.Context, isActive *bool, limit, offset int) ([]entity.Team, error) {
	teams, err := s.teamRepo.List(ctx, isActive, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	return teams, nil
}

// UpdateTeam updates a team
func (s *TeamService) UpdateTeam(ctx context.Context, id uuid.UUID, req entity.UpdateTeamRequest) (*entity.Team, error) {
	// Validate request
	if err := s.validate.Struct(req); err != nil {
		return nil, ParamsError{Message: fmt.Sprintf("validation failed: %v", err)}
	}

	// Get existing team
	team, err := s.teamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	// Update fields
	if req.Name != nil {
		// Check if new name conflicts with another team
		if *req.Name != team.Name {
			existingTeam, _ := s.teamRepo.GetByName(ctx, *req.Name)
			if existingTeam != nil {
				return nil, ParamsError{Message: "team with this name already exists"}
			}
		}
		team.Name = *req.Name
	}

	if req.Description != nil {
		team.Description = req.Description
	}

	if req.ManagerID != nil {
		// Verify manager exists
		manager, err := s.userRepo.GetByID(ctx, *req.ManagerID)
		if err != nil || manager == nil {
			return nil, ParamsError{Message: "manager not found"}
		}
		team.ManagerID = req.ManagerID
	}

	if req.IsActive != nil {
		team.IsActive = *req.IsActive
	}

	// Save changes
	if err := s.teamRepo.Update(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	return team, nil
}

// DeleteTeam soft deletes a team
func (s *TeamService) DeleteTeam(ctx context.Context, id uuid.UUID) error {
	if err := s.teamRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}

// AddTeamMember adds a member to a team
func (s *TeamService) AddTeamMember(ctx context.Context, teamID uuid.UUID, req entity.AddTeamMemberRequest) error {
	// Validate request
	if err := s.validate.Struct(req); err != nil {
		return ParamsError{Message: fmt.Sprintf("validation failed: %v", err)}
	}

	// Verify team exists
	_, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return ParamsError{Message: "team not found"}
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil || user == nil {
		return ParamsError{Message: "user not found"}
	}

	if !user.IsActive {
		return ParamsError{Message: "user is not active"}
	}

	// Check if already a member
	isMember, err := s.teamRepo.IsMember(ctx, teamID, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return ParamsError{Message: "user is already a team member"}
	}

	// Add member
	member := &entity.TeamMember{
		TeamID:   teamID,
		UserID:   req.UserID,
		Role:     req.Role,
		JoinedAt: time.Now(),
	}

	if err := s.teamRepo.AddMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add team member: %w", err)
	}

	return nil
}

// RemoveTeamMember removes a member from a team
func (s *TeamService) RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error {
	// Verify team exists
	_, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return ParamsError{Message: "team not found"}
	}

	// Remove member
	if err := s.teamRepo.RemoveMember(ctx, teamID, userID); err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	return nil
}

// GetTeamMembers retrieves all members of a team
func (s *TeamService) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]entity.TeamMember, error) {
	members, err := s.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	return members, nil
}

// TeamEnrichedInfo holds extra info about a team for list display.
type TeamEnrichedInfo struct {
	MemberCount int
	LeadName    string
}

// GetTeamEnrichedInfo returns member count and lead name for a team.
func (s *TeamService) GetTeamEnrichedInfo(ctx context.Context, team entity.Team) TeamEnrichedInfo {
	info := TeamEnrichedInfo{}
	members, err := s.teamRepo.GetMembers(ctx, team.ID)
	if err == nil {
		info.MemberCount = len(members)
	}
	if team.ManagerID != nil {
		if manager, err := s.userRepo.GetByID(ctx, *team.ManagerID); err == nil {
			info.LeadName = manager.FirstName + " " + manager.LastName
		}
	}
	return info
}

// UpdateMemberRole updates a team member's role
func (s *TeamService) UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	// Validate role
	validRoles := map[string]bool{"developer": true, "lead": true, "architect": true, "qa": true, "analyst": true, "devops": true, "sre": true}
	if !validRoles[role] {
		return ParamsError{Message: "invalid role. must be one of: developer, lead, architect, qa, analyst, devops, sre"}
	}

	if err := s.teamRepo.UpdateMemberRole(ctx, teamID, userID, role); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	return nil
}
