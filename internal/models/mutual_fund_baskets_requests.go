package models

// CreateMutualFundBasketRequest represents the request for creating a mutual fund basket
type CreateMutualFundBasketRequest struct {
	BasketType  string                               `json:"basketType" binding:"required"`
	Name        string                               `json:"name" binding:"required"`
	Description *string                              `json:"description"`
	MaxSchemes  int                                  `json:"maxSchemes"`
	Items       []*CreateMutualFundBasketItemRequest `json:"items" binding:"required"`
}

// CreateMutualFundBasketItemRequest represents a mutual fund basket item in create request
type CreateMutualFundBasketItemRequest struct {
	SchemeName   string   `json:"schemeName" binding:"required"`
	SchemeCode   *string  `json:"schemeCode"`
	EntryPrice   float64  `json:"entryPrice" binding:"required"`
	ExitPrice    *float64 `json:"exitPrice"`
	CurrentNAV   *float64 `json:"currentNav"`
	Trend        *string  `json:"trend"`
	DisplayOrder int      `json:"displayOrder"`
}

// GetMutualFundBasketParams represents path parameters for get mutual fund basket
type GetMutualFundBasketParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// UpdateMutualFundBasketParams represents path parameters for update mutual fund basket
type UpdateMutualFundBasketParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// UpdateMutualFundBasketRequest represents the request for updating a mutual fund basket
type UpdateMutualFundBasketRequest struct {
	Name   *string `json:"name"`
	Status *string `json:"status"`
}
