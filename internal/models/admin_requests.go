package models

// Stock list (master list) requests
type CreateStockRequest struct {
	Name     string  `json:"name" binding:"required"`
	Symbol   *string `json:"symbol"`
	Exchange *string `json:"exchange"`
}

type UpdateStockRequest struct {
	Name     *string `json:"name"`
	Symbol   *string `json:"symbol"`
	Exchange *string `json:"exchange"`
}

type GetStockParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// Stock Bullet Requests (stock_id preferred; name+exchange kept for backward compat)
type CreateStockBulletRequest struct {
	StockID   *string `json:"stockId"`
	Name      string  `json:"name"`
	Exchange  string  `json:"exchange"`
	Price     string  `json:"price" binding:"required"`
	Rationale *string `json:"rationale"`
	Verdict   string  `json:"verdict" binding:"required"`
}

type UpdateStockBulletRequest struct {
	StockID   *string `json:"stockId"`
	Name      *string `json:"name"`
	Exchange  *string `json:"exchange"`
	Price     *string `json:"price"`
	Rationale *string `json:"rationale"`
	Verdict   *string `json:"verdict"`
}

type GetStockBulletParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// Stock Recommendation Requests (stock_id preferred; name kept for backward compat)
type CreateStockRecommendationRequest struct {
	StockID *string `json:"stockId"`
	Name    string  `json:"name"`
	Price   string  `json:"price" binding:"required"`
	Verdict string  `json:"verdict" binding:"required"`
}

type UpdateStockRecommendationRequest struct {
	StockID *string `json:"stockId"`
	Name    *string `json:"name"`
	Price   *string `json:"price"`
	Verdict *string `json:"verdict"`
}

type GetStockRecommendationParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// MF Scheme Requests
type CreateMFSchemeRequest struct {
	Name        string  `json:"name" binding:"required"`
	SchemeCode  *string `json:"schemeCode"`
	AMC         string  `json:"amc" binding:"required"`
	Category    string  `json:"category" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	CurrentNAV  *float64 `json:"currentNAV"`
	EntryPrice  *float64 `json:"entryPrice"`
	ExitPrice   *float64 `json:"exitPrice"`
	Trend       *string  `json:"trend"`
	Verdict     *string  `json:"verdict"`
	Rationale   *string  `json:"rationale"`
	Status      string  `json:"status" binding:"required"`
}

type UpdateMFSchemeRequest struct {
	Name        *string  `json:"name"`
	SchemeCode  *string  `json:"schemeCode"`
	AMC         *string  `json:"amc"`
	Category    *string  `json:"category"`
	Type        *string  `json:"type"`
	CurrentNAV  *float64 `json:"currentNAV"`
	EntryPrice  *float64 `json:"entryPrice"`
	ExitPrice   *float64 `json:"exitPrice"`
	Trend       *string  `json:"trend"`
	Verdict     *string  `json:"verdict"`
	Rationale   *string  `json:"rationale"`
	Status      *string  `json:"status"`
}

// NFO Requests
type CreateNFORequest struct {
	Name              string   `json:"name" binding:"required"`
	AMC               string   `json:"amc" binding:"required"`
	Category          string   `json:"category" binding:"required"`
	Type              string   `json:"type" binding:"required"`
	OpenDate          *string  `json:"openDate"`
	CloseDate         *string  `json:"closeDate"`
	Timeline          *string  `json:"timeline"`
	Summary           []string `json:"summary"`
	MinimumInvestment *float64 `json:"minimumInvestment"`
	Verdict           *string  `json:"verdict"`
	Rationale         *string  `json:"rationale"`
	Status            string   `json:"status" binding:"required"`
}

type UpdateNFORequest struct {
	Name              *string   `json:"name"`
	AMC               *string   `json:"amc"`
	Category          *string   `json:"category"`
	Type              *string   `json:"type"`
	OpenDate          *string   `json:"openDate"`
	CloseDate         *string   `json:"closeDate"`
	Timeline          *string   `json:"timeline"`
	Summary           []string  `json:"summary"`
	MinimumInvestment *float64  `json:"minimumInvestment"`
	Verdict           *string   `json:"verdict"`
	Rationale         *string   `json:"rationale"`
	Status            *string   `json:"status"`
}

// Sector Requests
type CreateSectorRequest struct {
	SectorName      string   `json:"sectorName" binding:"required"`
	CurrentValue    *float64 `json:"currentValue"`
	ChangePercentage *float64 `json:"changePercentage"`
	ChangeValue     *float64 `json:"changeValue"`
	ReportURL       *string  `json:"reportURL"`
}

type UpdateSectorRequest struct {
	SectorName      *string   `json:"sectorName"`
	CurrentValue    *float64  `json:"currentValue"`
	ChangePercentage *float64 `json:"changePercentage"`
	ChangeValue     *float64  `json:"changeValue"`
	ReportURL       *string   `json:"reportURL"`
}

type GetSectorParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// Webinar Requests
type CreateWebinarRequest struct {
	Title          string  `json:"title" binding:"required"`
	Description    *string `json:"description"`
	Status         string  `json:"status" binding:"required"`
	Date           *string `json:"date"`
	Time           *string `json:"time"`
	Duration       *string `json:"duration"`
	RegistrationURL *string `json:"registrationURL"`
	RecordingURL   *string `json:"recordingURL"`
	ThumbnailURL   *string `json:"thumbnailURL"`
}

type UpdateWebinarRequest struct {
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	Status         *string `json:"status"`
	Date           *string `json:"date"`
	Time           *string `json:"time"`
	Duration       *string `json:"duration"`
	RegistrationURL *string `json:"registrationURL"`
	RecordingURL   *string `json:"recordingURL"`
	ThumbnailURL   *string `json:"thumbnailURL"`
}

// Market Pulse Requests
type CreateMarketPulseRequest struct {
	Content string `json:"content" binding:"required"`
	Updated string `json:"updated" binding:"required"`
}

type UpdateMarketPulseRequest struct {
	Content *string `json:"content"`
	Updated *string `json:"updated"`
}

type GetMarketPulseParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// Top Call Requests
type UpdateTopCallRequest struct {
	Type     string `json:"type" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Verdict  string `json:"verdict" binding:"required"`
	Rationale string `json:"rationale" binding:"required"`
}

// Weekly Market Mood Requests
type UpdateWeeklyMarketMoodRequest struct {
	Title         string   `json:"title" binding:"required"`
	Week          string   `json:"week" binding:"required"`
	Points        []string `json:"points"`
	Sentiment     float64  `json:"sentiment" binding:"required"`
	SentimentLabel string  `json:"sentimentLabel" binding:"required"`
	ReportURL     *string  `json:"reportURL"`
}

// Weekly Audio Requests
type UpdateWeeklyAudioRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	Duration    *string `json:"duration"`
	Week        *string `json:"week"`
	AudioURL    *string `json:"audioURL"`
	ThumbnailURL *string `json:"thumbnailURL"`
	Status      string  `json:"status" binding:"required"`
}
