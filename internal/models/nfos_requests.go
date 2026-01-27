package models

// GetNFOParams represents path parameters for get NFO
type GetNFOParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}
