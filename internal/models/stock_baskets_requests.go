package models

// CreateStockBasketRequest represents the request for creating a stock basket
type CreateStockBasketRequest struct {
	Name         string                          `json:"name" binding:"required"`
	Description  *string                         `json:"description"`
	IsBulletIdea bool                            `json:"isBulletIdea"`
	Items        []*CreateStockBasketItemRequest `json:"items" binding:"required"`
}

// CreateStockBasketItemRequest represents a stock basket item in create request (references stocks table)
type CreateStockBasketItemRequest struct {
	StockID         string   `json:"stockId" binding:"required"`
	CMP             float64  `json:"cmp" binding:"required"`
	Target          float64  `json:"target" binding:"required"`
	StopLoss        *float64 `json:"stopLoss"`
	EntryRangeMin   *float64 `json:"entryRangeMin"`
	EntryRangeMax   *float64 `json:"entryRangeMax"`
	Action          *string  `json:"action"`
	RiskLevel       *string  `json:"riskLevel"`
	TimeHorizon     *string  `json:"timeHorizon"`
	Rationale       *string  `json:"rationale"`
	ReportURL       *string  `json:"reportUrl"`
	FundamentalsURL *string  `json:"fundamentalsUrl"`
	DisplayOrder    int      `json:"displayOrder"`
	IsBulletIdea    bool     `json:"isBulletIdea"`
}

// GetStockBasketParams represents path parameters for get stock basket
type GetStockBasketParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// UpdateStockBasketParams represents path parameters for update stock basket
type UpdateStockBasketParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// UpdateStockBasketRequest represents the request for updating a stock basket
type UpdateStockBasketRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}
