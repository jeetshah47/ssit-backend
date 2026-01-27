package models

// CreateIPOAdvisoryRequest represents the request for creating an IPO advisory
type CreateIPOAdvisoryRequest struct {
	IPOName      string   `json:"ipoName" binding:"required"`
	IPOSymbol    *string  `json:"ipoSymbol"`
	GMP          *float64 `json:"gmp"`
	Suggestion   *string  `json:"suggestion"`
	LotSize      int      `json:"lotSize" binding:"required"`
	PriceBandMin float64  `json:"priceBandMin" binding:"required"`
	PriceBandMax float64  `json:"priceBandMax" binding:"required"`
	IssueDate    *string  `json:"issueDate"`
	IssueSize    *string  `json:"issueSize"`
	IPOTimetable *string  `json:"ipoTimetable"`
}

// GetIPOAdvisoryParams represents path parameters for get IPO advisory
type GetIPOAdvisoryParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// UpdateIPOAdvisoryParams represents path parameters for update IPO advisory
type UpdateIPOAdvisoryParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// UpdateIPOAdvisoryRequest represents the request for updating an IPO advisory
type UpdateIPOAdvisoryRequest struct {
	IPOName      *string  `json:"ipoName"`
	GMP          *float64 `json:"gmp"`
	Suggestion   *string  `json:"suggestion"`
	Status       *string  `json:"status"`
	IssueDate    *string  `json:"issueDate"`
	IPOTimetable *string  `json:"ipoTimetable"`
}
