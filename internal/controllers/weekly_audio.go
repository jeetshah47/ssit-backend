package controllers

import (
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/google/uuid"
)

// WeeklyAudioController handles weekly audio endpoints
type WeeklyAudioController struct {
	getWeeklyAudioService   *services.GetWeeklyAudioService
	listWeeklyAudioService *services.ListWeeklyAudioService
}

// NewWeeklyAudioController creates a new weekly audio controller
func NewWeeklyAudioController(
	getWeeklyAudioService *services.GetWeeklyAudioService,
	listWeeklyAudioService *services.ListWeeklyAudioService,
) *WeeklyAudioController {
	return &WeeklyAudioController{
		getWeeklyAudioService:   getWeeklyAudioService,
		listWeeklyAudioService: listWeeklyAudioService,
	}
}

// GetWeeklyAudio retrieves weekly audio (latest if no ID provided)
func (c *WeeklyAudioController) GetWeeklyAudio(ctx *utils.Context) (interface{}, error) {
	var params models.GetWeeklyAudioParams
	// ID is optional - if not provided, get latest
	audioID := uuid.Nil
	if err := ctx.BindURI(&params); err == nil && params.ID != "" {
		var parseErr error
		audioID, parseErr = uuid.Parse(params.ID)
		if parseErr != nil {
			return nil, parseErr
		}
	}

	query := services.GetWeeklyAudioQuery{
		AudioID: audioID,
	}

	result, err := c.getWeeklyAudioService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Audio, nil
}

// ListWeeklyAudio lists weekly audio
func (c *WeeklyAudioController) ListWeeklyAudio(ctx *utils.Context) (interface{}, error) {
	query := services.ListWeeklyAudioQuery{
		Limit:  10,
		Offset: 0,
	}

	if limit := ctx.GetQuery("limit"); limit != "" {
		if parsedLimit := parseInt(limit); parsedLimit > 0 {
			query.Limit = parsedLimit
		}
	}
	if offset := ctx.GetQuery("offset"); offset != "" {
		if parsedOffset := parseInt(offset); parsedOffset >= 0 {
			query.Offset = parsedOffset
		}
	}

	result, err := c.listWeeklyAudioService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Audios, nil
}
