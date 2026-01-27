package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WeeklyAudioRepo defines the interface for weekly audio data persistence
type WeeklyAudioRepo interface {
	Create(ctx context.Context, audio *models.WeeklyAudio) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.WeeklyAudio, error)
	FindLatest(ctx context.Context) (*models.WeeklyAudio, error)
	FindAll(ctx context.Context, limit, offset int) ([]*models.WeeklyAudio, error)
	Update(ctx context.Context, audio *models.WeeklyAudio) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// weeklyAudioRepo implements the weekly audio repository interface
type weeklyAudioRepo struct {
	*RepoContext
}

// NewWeeklyAudioRepo creates a new weekly audio repository
func NewWeeklyAudioRepo(ctx *RepoContext) WeeklyAudioRepo {
	return &weeklyAudioRepo{RepoContext: ctx}
}

// Create creates a new weekly audio
func (r *weeklyAudioRepo) Create(ctx context.Context, audio *models.WeeklyAudio) error {
	if err := r.db.WithContext(ctx).Create(audio).Error; err != nil {
		return fmt.Errorf("failed to create weekly audio: %w", err)
	}
	return nil
}

// FindByID finds a weekly audio by ID
func (r *weeklyAudioRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.WeeklyAudio, error) {
	var audio models.WeeklyAudio
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&audio).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("WEEKLY_AUDIO_NOT_FOUND", "weekly audio not found")
		}
		return nil, fmt.Errorf("failed to find weekly audio: %w", err)
	}
	return &audio, nil
}

// FindLatest finds the latest weekly audio
func (r *weeklyAudioRepo) FindLatest(ctx context.Context) (*models.WeeklyAudio, error) {
	var audio models.WeeklyAudio
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		First(&audio).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("WEEKLY_AUDIO_NOT_FOUND", "no weekly audio found")
		}
		return nil, fmt.Errorf("failed to find latest weekly audio: %w", err)
	}
	return &audio, nil
}

// FindAll finds all weekly audio
func (r *weeklyAudioRepo) FindAll(ctx context.Context, limit, offset int) ([]*models.WeeklyAudio, error) {
	var audios []*models.WeeklyAudio
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&audios).Error; err != nil {
		return nil, fmt.Errorf("failed to find weekly audio: %w", err)
	}
	return audios, nil
}

// Update updates an existing weekly audio
func (r *weeklyAudioRepo) Update(ctx context.Context, audio *models.WeeklyAudio) error {
	if err := r.db.WithContext(ctx).Save(audio).Error; err != nil {
		return fmt.Errorf("failed to update weekly audio: %w", err)
	}
	return nil
}

// Delete deletes a weekly audio
func (r *weeklyAudioRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.WeeklyAudio{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete weekly audio: %w", err)
	}
	return nil
}
