package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// RoleRepo defines the interface for role data persistence
type RoleRepo interface {
	// FindUserRoles finds all role names for a user
	FindUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error)

	// UserHasRole checks if a user has a specific role
	UserHasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error)
}

// roleRepo implements the role repository interface
type roleRepo struct {
	*RepoContext
}

// NewRoleRepo creates a new role repository
func NewRoleRepo(ctx *RepoContext) RoleRepo {
	return &roleRepo{RepoContext: ctx}
}

// FindUserRoles finds all role names for a user by joining roles and user_roles tables
func (r *roleRepo) FindUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var roleNames []string

	// SQL query to get role names for a user
	// SELECT r.name FROM roles r
	// JOIN user_roles ur ON r.id = ur.role_id
	// WHERE ur.user_id = ?
	result := r.db.WithContext(ctx).
		Table("roles").
		Select("roles.name").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Pluck("roles.name", &roleNames)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user roles: %w", result.Error)
	}

	// If no roles found, return empty slice (not nil)
	if roleNames == nil {
		return []string{}, nil
	}

	return roleNames, nil
}

// UserHasRole checks if a user has a specific role (case-insensitive match).
func (r *roleRepo) UserHasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error) {
	var count int64

	// SQL: match role name case-insensitively so "admin" matches "Admin" in DB
	res := r.db.WithContext(ctx).
		Table("user_roles").
		Select("COUNT(*)").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND LOWER(TRIM(roles.name)) = LOWER(TRIM(?))", userID, roleName).
		Count(&count)

	if res.Error != nil {
		return false, fmt.Errorf("failed to check user role: %w", res.Error)
	}

	return count > 0, nil
}
