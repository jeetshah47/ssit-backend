package models

// GetMFSchemeParams represents path parameters for get MF scheme
type GetMFSchemeParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}
