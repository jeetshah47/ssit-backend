package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/equitywala/backend/internal/domain/user"
)

// GetUserQuery represents the input for getting a user
type GetUserQuery struct {
	UserID uuid.UUID
}

// GetUserResult represents the output
type GetUserResult struct {
	User *user.User
}

// GetUserHandler handles user retrieval
type GetUserHandler struct {
	userRepo user.Repository
}

// NewGetUserHandler creates a new query handler
func NewGetUserHandler(userRepo user.Repository) *GetUserHandler {
	return &GetUserHandler{
		userRepo: userRepo,
	}
}

// Handle executes the query
func (h *GetUserHandler) Handle(ctx context.Context, query GetUserQuery) (*GetUserResult, error) {
	user, err := h.userRepo.FindByID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}

	return &GetUserResult{User: user}, nil
}

