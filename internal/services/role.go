package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// RoleService handles role-related business logic
type RoleService struct {
	roleRepo repositories.RoleRepo
}

// NewRoleService creates a new role service
func NewRoleService(roleRepo repositories.RoleRepo) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
	}
}

// GetUserRoles returns all role names for a user (normalized to lowercase for consistent mapping).
func (s *RoleService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	roles, err := s.roleRepo.FindUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// If no roles found, return default "user" role for backward compatibility
	if len(roles) == 0 {
		return []string{"user"}, nil
	}

	// Normalize role names to lowercase so "Admin" in DB maps to "admin" for API and frontend
	out := make([]string, len(roles))
	for i, r := range roles {
		out[i] = strings.ToLower(strings.TrimSpace(r))
	}
	return out, nil
}

// UserHasRole checks if a user has a specific role
func (s *RoleService) UserHasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error) {
	hasRole, err := s.roleRepo.UserHasRole(ctx, userID, roleName)
	if err != nil {
		return false, fmt.Errorf("failed to check user role: %w", err)
	}

	return hasRole, nil
}

// GetPrimaryRole returns the highest privilege role for a user
// Priority: admin > advisor > user
func (s *RoleService) GetPrimaryRole(ctx context.Context, userID uuid.UUID) (string, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return "user", err
	}

	// Check for admin role first (highest privilege)
	for _, role := range roles {
		if role == "admin" {
			return "admin", nil
		}
	}

	// Check for advisor role
	for _, role := range roles {
		if role == "advisor" {
			return "advisor", nil
		}
	}

	// Default to user role
	return "user", nil
}
