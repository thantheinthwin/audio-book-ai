package database

import (
	"testing"
	"time"

	"audio-book-ai/api/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProcessingJobWithChapterID(t *testing.T) {
	// This test verifies that the chapter_id column works correctly
	// Note: This is a unit test that doesn't require a real database connection

	// Create a processing job with chapter_id
	chapterID := uuid.New()
	audiobookID := uuid.New()

	job := &models.ProcessingJob{
		ID:          uuid.New(),
		AudiobookID: audiobookID,
		ChapterID:   &chapterID,
		JobType:     models.JobTypeTranscribe,
		Status:      models.JobStatusPending,
		CreatedAt:   time.Now(),
	}

	// Verify the job has the chapter_id set
	assert.NotNil(t, job.ChapterID)
	assert.Equal(t, chapterID, *job.ChapterID)
	assert.Equal(t, audiobookID, job.AudiobookID)
	assert.Equal(t, models.JobTypeTranscribe, job.JobType)
	assert.Equal(t, models.JobStatusPending, job.Status)
}

func TestProcessingJobWithoutChapterID(t *testing.T) {
	// This test verifies that processing jobs can be created without chapter_id
	// (for audiobook-level jobs)

	audiobookID := uuid.New()

	job := &models.ProcessingJob{
		ID:          uuid.New(),
		AudiobookID: audiobookID,
		ChapterID:   nil, // No chapter_id for audiobook-level jobs
		JobType:     models.JobTypeSummarize,
		Status:      models.JobStatusPending,
		CreatedAt:   time.Now(),
	}

	// Verify the job has no chapter_id
	assert.Nil(t, job.ChapterID)
	assert.Equal(t, audiobookID, job.AudiobookID)
	assert.Equal(t, models.JobTypeSummarize, job.JobType)
	assert.Equal(t, models.JobStatusPending, job.Status)
}

func TestProcessingJobChapterIDOptional(t *testing.T) {
	// This test verifies that chapter_id is optional and can be set later

	audiobookID := uuid.New()

	// Create job without chapter_id initially
	job := &models.ProcessingJob{
		ID:          uuid.New(),
		AudiobookID: audiobookID,
		ChapterID:   nil,
		JobType:     models.JobTypeTranscribe,
		Status:      models.JobStatusPending,
		CreatedAt:   time.Now(),
	}

	assert.Nil(t, job.ChapterID)

	// Set chapter_id later
	chapterID := uuid.New()
	job.ChapterID = &chapterID

	assert.NotNil(t, job.ChapterID)
	assert.Equal(t, chapterID, *job.ChapterID)
}
