package models

// CreateETFBasketRequest represents the request for creating an ETF basket
type CreateETFBasketRequest struct {
	Name        string                      `json:"name" binding:"required"`
	Description *string                     `json:"description"`
	Items       []*CreateETFBasketItemRequest `json:"items"`
}

// CreateETFBasketItemRequest represents an ETF basket item in create request
type CreateETFBasketItemRequest struct {
	Name          string   `json:"name" binding:"required"`
	Symbol        *string  `json:"symbol"`
	CMP           float64  `json:"cmp" binding:"required"`
	Target        float64  `json:"target" binding:"required"`
	StopLoss      *float64 `json:"stopLoss"`
	EntryRangeMin *float64 `json:"entryRangeMin"`
	EntryRangeMax *float64 `json:"entryRangeMax"`
	Action        *string  `json:"action"`
	RiskLevel     *string  `json:"riskLevel"`
	TimeHorizon   *string  `json:"timeHorizon"`
	Rationale     *string  `json:"rationale"`
	DisplayOrder  int      `json:"displayOrder"`
}

// GetETFBasketParams represents path parameters for get ETF basket
type GetETFBasketParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// UpdateETFBasketParams represents path parameters for update ETF basket
type UpdateETFBasketParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// UpdateETFBasketRequest represents the request for updating an ETF basket
type UpdateETFBasketRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}
