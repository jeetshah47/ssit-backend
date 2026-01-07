package user

import "github.com/equitywala/backend/internal/shared/errors"

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.NewDomainError(
		"USER_NOT_FOUND",
		"user not found",
	)

	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.NewDomainError(
		"INVALID_EMAIL",
		"invalid email format",
	)

	// ErrUserAlreadyExists is returned when trying to create a user that already exists
	ErrUserAlreadyExists = errors.NewDomainError(
		"USER_ALREADY_EXISTS",
		"user with this email already exists",
	)

	// ErrInvalidPassword is returned when password is invalid
	ErrInvalidPassword = errors.NewDomainError(
		"INVALID_PASSWORD",
		"invalid password",
	)

	// ErrUserNotActive is returned when user account is not active
	ErrUserNotActive = errors.NewDomainError(
		"USER_NOT_ACTIVE",
		"user account is not active",
	)
)

