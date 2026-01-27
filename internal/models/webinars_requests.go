package models

// GetWebinarParams represents path parameters for get webinar
type GetWebinarParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}
