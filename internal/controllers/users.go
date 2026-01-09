package controllers

import (
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/equitywala/backend/internal/interfaces/http/api"
	"github.com/google/uuid"
)

// UserController handles user endpoints
type UserController struct {
	getUserService *services.GetUserService
}

// NewUserController creates a new user controller
func NewUserController(getUserService *services.GetUserService) *UserController {
	return &UserController{
		getUserService: getUserService,
	}
}

// GetUser retrieves a user by ID
func (c *UserController) GetUser(ctx *api.Context) (interface{}, error) {
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

	return toUserResponse(result.User), nil
}

