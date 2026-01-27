package models

// UploadReportParams represents path parameters for upload report
type UploadReportParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// DeleteReportParams represents path parameters for delete report
type DeleteReportParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// DeleteReportRequest represents the request for deleting a report
type DeleteReportRequest struct {
	ReportURL string `json:"reportUrl" binding:"required"`
}
