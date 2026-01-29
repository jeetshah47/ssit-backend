package controllers

import (
	"context"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/google/uuid"
)

// UserController handles user endpoints
type UserController struct {
	getUserService *services.GetUserService
	roleService    *services.RoleService
}

// NewUserController creates a new user controller
func NewUserController(getUserService *services.GetUserService, roleService *services.RoleService) *UserController {
	return &UserController{
		getUserService: getUserService,
		roleService:    roleService,
	}
}

// GetUser retrieves a user by ID
func (c *UserController) GetUser(ctx *utils.Context) (interface{}, error) {
	var params models.GetUserParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	query := services.GetUserQuery{
		UserID: userID,
	}

	result, err := c.getUserService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	// Convert to UserResponse with roles
	return c.toUserResponse(ctx.Request.Context(), result.User), nil
}

// toUserResponse converts a User model to UserResponse with roles
func (c *UserController) toUserResponse(ctx context.Context, u *models.User) *models.UserResponse {
	userResponse := &models.UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Phone:         u.Phone,
		Name:          u.Name,
		Status:        u.Status,
		IsMfCustomer:  u.GetIsMfCustomer(),
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Get user roles if role service is available
	if c.roleService != nil {
		roles, err := c.roleService.GetUserRoles(ctx, u.ID)
		if err == nil {
			userResponse.Roles = roles

			// Get primary role
			primaryRole, err := c.roleService.GetPrimaryRole(ctx, u.ID)
			if err == nil {
				userResponse.PrimaryRole = primaryRole
			} else {
				// Default to user if error
				userResponse.PrimaryRole = "user"
			}
		} else {
			// If error fetching roles, default to user
			userResponse.Roles = []string{"user"}
			userResponse.PrimaryRole = "user"
		}
	} else {
		// If role service not available, default to user
		userResponse.Roles = []string{"user"}
		userResponse.PrimaryRole = "user"
	}

	return userResponse
}

