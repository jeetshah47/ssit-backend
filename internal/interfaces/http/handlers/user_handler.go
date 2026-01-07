package handlers

import (
	"time"

	"github.com/equitywala/backend/internal/application/user/queries"
	"github.com/equitywala/backend/internal/domain/user"
	"github.com/equitywala/backend/internal/interfaces/http/api"
	"github.com/equitywala/backend/internal/interfaces/http/dto"
	"github.com/google/uuid"
)

// UserHandler handles user endpoints
type UserHandler struct {
	getUserQuery *queries.GetUserHandler
}

// NewUserHandler creates a new user handler
func NewUserHandler(getUserQuery *queries.GetUserHandler) *UserHandler {
	return &UserHandler{
		getUserQuery: getUserQuery,
	}
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(ctx *api.Context) (interface{}, error) {
	var params dto.GetUserParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	query := queries.GetUserQuery{
		UserID: userID,
	}

	result, err := h.getUserQuery.Handle(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return toUserResponseDTO(result.User), nil
}

func toUserResponseDTO(u *user.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Phone:         u.Phone,
		Name:          u.Name,
		Status:        u.Status,
		IsMfCustomer:   u.GetIsMfCustomer(),
		EmailVerified: u.EmailVerified,
		CreatedAt:      u.CreatedAt.Format(time.RFC3339),
	}
}

