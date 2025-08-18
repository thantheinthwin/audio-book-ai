package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"audio-book-ai/api/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockRepository is a mock implementation of the Repository interface for testing
type MockRepository struct {
	audiobooks     map[uuid.UUID]*models.AudioBook
	transcripts    map[uuid.UUID]*models.Transcript
	aiOutputs      map[uuid.UUID]*models.AIOutput
	processingJobs map[uuid.UUID]*models.ProcessingJob
	tags           map[uuid.UUID]*models.Tag
	audiobookTags  map[string]*models.AudioBookTag
	embeddings     map[uuid.UUID]*models.AudioBookEmbedding
}

// NewMockRepository creates a new mock repository for testing
func NewMockRepository() *MockRepository {
	return &MockRepository{
		audiobooks:     make(map[uuid.UUID]*models.AudioBook),
		transcripts:    make(map[uuid.UUID]*models.Transcript),
		aiOutputs:      make(map[uuid.UUID]*models.AIOutput),
		processingJobs: make(map[uuid.UUID]*models.ProcessingJob),
		tags:           make(map[uuid.UUID]*models.Tag),
		audiobookTags:  make(map[string]*models.AudioBookTag),
		embeddings:     make(map[uuid.UUID]*models.AudioBookEmbedding),
	}
}

// AudioBook operations
func (m *MockRepository) CreateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
	if audiobook.ID == uuid.Nil {
		audiobook.ID = uuid.New()
	}
	m.audiobooks[audiobook.ID] = audiobook
	return nil
}

func (m *MockRepository) GetAudioBookByID(ctx context.Context, id uuid.UUID) (*models.AudioBook, error) {
	if audiobook, exists := m.audiobooks[id]; exists {
		return audiobook, nil
	}
	return nil, ErrNotFound
}

func (m *MockRepository) GetAudioBookWithDetails(ctx context.Context, id uuid.UUID) (*models.AudioBookWithDetails, error) {
	audiobook, err := m.GetAudioBookByID(ctx, id)
	if err != nil {
		return nil, err
	}

	details := &models.AudioBookWithDetails{
		AudioBook: *audiobook,
	}

	// Get transcript
	transcript, _ := m.GetTranscriptByAudioBookID(ctx, id)
	details.Transcript = transcript

	// Get AI outputs
	aiOutputs, _ := m.GetAIOutputsByAudioBookID(ctx, id)
	details.AIOutputs = aiOutputs

	// Get tags
	tags, _ := m.GetTagsByAudioBookID(ctx, id)
	details.Tags = tags

	// Get processing jobs
	jobs, _ := m.GetProcessingJobsByAudioBookID(ctx, id)
	details.ProcessingJobs = jobs

	return details, nil
}

func (m *MockRepository) UpdateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
	if _, exists := m.audiobooks[audiobook.ID]; !exists {
		return ErrNotFound
	}
	m.audiobooks[audiobook.ID] = audiobook
	return nil
}

func (m *MockRepository) DeleteAudioBook(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.audiobooks[id]; !exists {
		return ErrNotFound
	}
	delete(m.audiobooks, id)
	return nil
}

func (m *MockRepository) ListAudioBooks(ctx context.Context, limit, offset int, isPublic *bool) ([]models.AudioBook, int, error) {
	var audiobooks []models.AudioBook
	for _, audiobook := range m.audiobooks {
		if isPublic == nil || audiobook.IsPublic == *isPublic {
			audiobooks = append(audiobooks, *audiobook)
		}
	}
	total := len(audiobooks)

	// Simple pagination
	if offset >= total {
		return []models.AudioBook{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return audiobooks[offset:end], total, nil
}

func (m *MockRepository) GetAudioBooksByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error) {
	var audiobooks []models.AudioBook
	for _, audiobook := range m.audiobooks {
		if audiobook.CreatedBy == userID {
			audiobooks = append(audiobooks, *audiobook)
		}
	}
	total := len(audiobooks)

	// Simple pagination
	if offset >= total {
		return []models.AudioBook{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return audiobooks[offset:end], total, nil
}

func (m *MockRepository) UpdateAudioBookStatus(ctx context.Context, id uuid.UUID, status models.AudioBookStatus) error {
	audiobook, err := m.GetAudioBookByID(ctx, id)
	if err != nil {
		return err
	}
	audiobook.Status = status
	return m.UpdateAudioBook(ctx, audiobook)
}

// Transcript operations
func (m *MockRepository) CreateTranscript(ctx context.Context, transcript *models.Transcript) error {
	if transcript.ID == uuid.Nil {
		transcript.ID = uuid.New()
	}
	m.transcripts[transcript.ID] = transcript
	return nil
}

func (m *MockRepository) GetTranscriptByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Transcript, error) {
	for _, transcript := range m.transcripts {
		if transcript.AudiobookID == audiobookID {
			return transcript, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateTranscript(ctx context.Context, transcript *models.Transcript) error {
	if _, exists := m.transcripts[transcript.ID]; !exists {
		return ErrNotFound
	}
	m.transcripts[transcript.ID] = transcript
	return nil
}

func (m *MockRepository) DeleteTranscript(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.transcripts[id]; !exists {
		return ErrNotFound
	}
	delete(m.transcripts, id)
	return nil
}

// AI Output operations
func (m *MockRepository) CreateAIOutput(ctx context.Context, output *models.AIOutput) error {
	if output.ID == uuid.Nil {
		output.ID = uuid.New()
	}
	m.aiOutputs[output.ID] = output
	return nil
}

func (m *MockRepository) GetAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AIOutput, error) {
	var outputs []models.AIOutput
	for _, output := range m.aiOutputs {
		if output.AudiobookID == audiobookID {
			outputs = append(outputs, *output)
		}
	}
	return outputs, nil
}

func (m *MockRepository) GetAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.AIOutput, error) {
	for _, output := range m.aiOutputs {
		if output.AudiobookID == audiobookID && output.OutputType == outputType {
			return output, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateAIOutput(ctx context.Context, output *models.AIOutput) error {
	if _, exists := m.aiOutputs[output.ID]; !exists {
		return ErrNotFound
	}
	m.aiOutputs[output.ID] = output
	return nil
}

func (m *MockRepository) DeleteAIOutput(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.aiOutputs[id]; !exists {
		return ErrNotFound
	}
	delete(m.aiOutputs, id)
	return nil
}

// Processing Job operations
func (m *MockRepository) CreateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	m.processingJobs[job.ID] = job
	return nil
}

func (m *MockRepository) GetProcessingJobsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ProcessingJob, error) {
	var jobs []models.ProcessingJob
	for _, job := range m.processingJobs {
		if job.AudiobookID == audiobookID {
			jobs = append(jobs, *job)
		}
	}
	return jobs, nil
}

func (m *MockRepository) GetProcessingJobByID(ctx context.Context, id uuid.UUID) (*models.ProcessingJob, error) {
	if job, exists := m.processingJobs[id]; exists {
		return job, nil
	}
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	if _, exists := m.processingJobs[job.ID]; !exists {
		return ErrNotFound
	}
	m.processingJobs[job.ID] = job
	return nil
}

func (m *MockRepository) GetPendingJobs(ctx context.Context, jobType models.JobType, limit int) ([]models.ProcessingJob, error) {
	var jobs []models.ProcessingJob
	for _, job := range m.processingJobs {
		if job.JobType == jobType && job.Status == models.JobStatusPending {
			jobs = append(jobs, *job)
		}
	}
	if len(jobs) > limit {
		jobs = jobs[:limit]
	}
	return jobs, nil
}

func (m *MockRepository) GetJobsByStatus(ctx context.Context, status models.JobStatus, limit int) ([]models.ProcessingJob, error) {
	var jobs []models.ProcessingJob
	for _, job := range m.processingJobs {
		if job.Status == status {
			jobs = append(jobs, *job)
		}
	}
	if len(jobs) > limit {
		jobs = jobs[:limit]
	}
	return jobs, nil
}

// Tag operations
func (m *MockRepository) CreateTag(ctx context.Context, tag *models.Tag) error {
	if tag.ID == uuid.Nil {
		tag.ID = uuid.New()
	}
	m.tags[tag.ID] = tag
	return nil
}

func (m *MockRepository) GetTagByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	if tag, exists := m.tags[id]; exists {
		return tag, nil
	}
	return nil, ErrNotFound
}

func (m *MockRepository) GetTagByName(ctx context.Context, name string) (*models.Tag, error) {
	for _, tag := range m.tags {
		if tag.Name == name {
			return tag, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) GetTagsByCategory(ctx context.Context, category string) ([]models.Tag, error) {
	var tags []models.Tag
	for _, tag := range m.tags {
		if tag.Category != nil && *tag.Category == category {
			tags = append(tags, *tag)
		}
	}
	return tags, nil
}

func (m *MockRepository) UpdateTag(ctx context.Context, tag *models.Tag) error {
	if _, exists := m.tags[tag.ID]; !exists {
		return ErrNotFound
	}
	m.tags[tag.ID] = tag
	return nil
}

func (m *MockRepository) DeleteTag(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.tags[id]; !exists {
		return ErrNotFound
	}
	delete(m.tags, id)
	return nil
}

func (m *MockRepository) ListTags(ctx context.Context, limit, offset int) ([]models.Tag, int, error) {
	var tags []models.Tag
	for _, tag := range m.tags {
		tags = append(tags, *tag)
	}
	total := len(tags)

	// Simple pagination
	if offset >= total {
		return []models.Tag{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return tags[offset:end], total, nil
}

// AudioBook Tag operations
func (m *MockRepository) CreateAudioBookTag(ctx context.Context, audiobookTag *models.AudioBookTag) error {
	key := audiobookTag.AudiobookID.String() + ":" + audiobookTag.TagID.String()
	m.audiobookTags[key] = audiobookTag
	return nil
}

func (m *MockRepository) GetTagsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Tag, error) {
	var tags []models.Tag
	for _, audiobookTag := range m.audiobookTags {
		if audiobookTag.AudiobookID == audiobookID {
			tagID := audiobookTag.TagID
			if tag, exists := m.tags[tagID]; exists {
				tags = append(tags, *tag)
			}
		}
	}
	return tags, nil
}

func (m *MockRepository) GetAudioBooksByTagID(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error) {
	var audiobooks []models.AudioBook
	for _, audiobookTag := range m.audiobookTags {
		if audiobookTag.TagID == tagID {
			if audiobook, exists := m.audiobooks[audiobookTag.AudiobookID]; exists {
				audiobooks = append(audiobooks, *audiobook)
			}
		}
	}
	total := len(audiobooks)

	// Simple pagination
	if offset >= total {
		return []models.AudioBook{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return audiobooks[offset:end], total, nil
}

func (m *MockRepository) DeleteAudioBookTag(ctx context.Context, audiobookID, tagID uuid.UUID) error {
	key := audiobookID.String() + ":" + tagID.String()
	if _, exists := m.audiobookTags[key]; !exists {
		return ErrNotFound
	}
	delete(m.audiobookTags, key)
	return nil
}

func (m *MockRepository) DeleteAllAudioBookTags(ctx context.Context, audiobookID uuid.UUID) error {
	for key, audiobookTag := range m.audiobookTags {
		if audiobookTag.AudiobookID == audiobookID {
			delete(m.audiobookTags, key)
		}
	}
	return nil
}

// AudioBook Embedding operations
func (m *MockRepository) CreateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error {
	if embedding.ID == uuid.Nil {
		embedding.ID = uuid.New()
	}
	m.embeddings[embedding.ID] = embedding
	return nil
}

func (m *MockRepository) GetEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AudioBookEmbedding, error) {
	var embeddings []models.AudioBookEmbedding
	for _, embedding := range m.embeddings {
		if embedding.AudiobookID == audiobookID {
			embeddings = append(embeddings, *embedding)
		}
	}
	return embeddings, nil
}

func (m *MockRepository) GetEmbeddingByType(ctx context.Context, audiobookID uuid.UUID, embeddingType models.EmbeddingType) (*models.AudioBookEmbedding, error) {
	for _, embedding := range m.embeddings {
		if embedding.AudiobookID == audiobookID && embedding.EmbeddingType == embeddingType {
			return embedding, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error {
	if _, exists := m.embeddings[embedding.ID]; !exists {
		return ErrNotFound
	}
	m.embeddings[embedding.ID] = embedding
	return nil
}

func (m *MockRepository) DeleteAudioBookEmbedding(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.embeddings[id]; !exists {
		return ErrNotFound
	}
	delete(m.embeddings, id)
	return nil
}

func (m *MockRepository) DeleteEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	for id, embedding := range m.embeddings {
		if embedding.AudiobookID == audiobookID {
			delete(m.embeddings, id)
		}
	}
	return nil
}

// Search operations
func (m *MockRepository) SearchAudioBooks(ctx context.Context, query string, limit, offset int, language *string, isPublic *bool) ([]models.AudioBook, int, error) {
	var audiobooks []models.AudioBook
	for _, audiobook := range m.audiobooks {
		// Simple text search
		if contains(audiobook.Title, query) || contains(audiobook.Author, query) {
			if language == nil || audiobook.Language == *language {
				if isPublic == nil || audiobook.IsPublic == *isPublic {
					audiobooks = append(audiobooks, *audiobook)
				}
			}
		}
	}
	total := len(audiobooks)

	// Simple pagination
	if offset >= total {
		return []models.AudioBook{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return audiobooks[offset:end], total, nil
}

func (m *MockRepository) SearchAudioBooksByVector(ctx context.Context, embedding []float64, embeddingType models.EmbeddingType, limit, offset int) ([]models.AudioBook, []float64, error) {
	// Mock implementation - just return all audiobooks with random similarity scores
	var audiobooks []models.AudioBook
	var similarities []float64

	for _, audiobook := range m.audiobooks {
		audiobooks = append(audiobooks, *audiobook)
		similarities = append(similarities, 0.8) // Mock similarity score
	}

	total := len(audiobooks)
	if offset >= total {
		return []models.AudioBook{}, []float64{}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return audiobooks[offset:end], similarities[offset:end], nil
}

func (m *MockRepository) SearchAudioBooksByTags(ctx context.Context, tagNames []string, limit, offset int) ([]models.AudioBook, int, error) {
	var audiobooks []models.AudioBook
	audiobookSet := make(map[uuid.UUID]bool)

	for _, tagName := range tagNames {
		for _, tag := range m.tags {
			if tag.Name == tagName {
				for _, audiobookTag := range m.audiobookTags {
					if audiobookTag.TagID == tag.ID {
						if audiobook, exists := m.audiobooks[audiobookTag.AudiobookID]; exists {
							audiobookSet[audiobook.ID] = true
						}
					}
				}
			}
		}
	}

	for audiobookID := range audiobookSet {
		if audiobook, exists := m.audiobooks[audiobookID]; exists {
			audiobooks = append(audiobooks, *audiobook)
		}
	}

	total := len(audiobooks)
	if offset >= total {
		return []models.AudioBook{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return audiobooks[offset:end], total, nil
}

// Utility operations
func (m *MockRepository) GetAudioBookStats(ctx context.Context) (*AudioBookStats, error) {
	stats := &AudioBookStats{}

	for _, audiobook := range m.audiobooks {
		stats.TotalAudioBooks++

		switch audiobook.Status {
		case models.StatusCompleted:
			stats.CompletedAudioBooks++
		case models.StatusPending:
			stats.PendingAudioBooks++
		case models.StatusProcessing:
			stats.ProcessingAudioBooks++
		case models.StatusFailed:
			stats.FailedAudioBooks++
		}

		if audiobook.DurationSeconds != nil {
			stats.TotalDuration += *audiobook.DurationSeconds
		}
		if audiobook.FileSizeBytes != nil {
			stats.TotalFileSize += *audiobook.FileSizeBytes
		}
	}

	return stats, nil
}

func (m *MockRepository) GetUserAudioBookStats(ctx context.Context, userID uuid.UUID) (*UserAudioBookStats, error) {
	stats := &UserAudioBookStats{UserID: userID}

	for _, audiobook := range m.audiobooks {
		if audiobook.CreatedBy == userID {
			stats.TotalAudioBooks++

			switch audiobook.Status {
			case models.StatusCompleted:
				stats.CompletedAudioBooks++
			case models.StatusPending:
				stats.PendingAudioBooks++
			case models.StatusProcessing:
				stats.ProcessingAudioBooks++
			case models.StatusFailed:
				stats.FailedAudioBooks++
			}

			if audiobook.DurationSeconds != nil {
				stats.TotalDuration += *audiobook.DurationSeconds
			}
			if audiobook.FileSizeBytes != nil {
				stats.TotalFileSize += *audiobook.FileSizeBytes
			}
		}
	}

	return stats, nil
}

func (m *MockRepository) CleanupOrphanedData(ctx context.Context) error {
	// Mock implementation - just return success
	return nil
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0)
}

// Error definitions
var ErrNotFound = &NotFoundError{}

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "not found"
}

// Tests
func TestMockRepository_CreateAudioBook(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	audiobook := &models.AudioBook{
		Title:     "Test Book",
		Author:    "Test Author",
		FilePath:  "/path/to/file.mp3",
		Language:  "en",
		Status:    models.StatusPending,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.CreateAudioBook(ctx, audiobook)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, audiobook.ID)

	// Verify it was created
	retrieved, err := repo.GetAudioBookByID(ctx, audiobook.ID)
	assert.NoError(t, err)
	assert.Equal(t, audiobook.Title, retrieved.Title)
}

func TestMockRepository_GetAudioBookByID_NotFound(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	_, err := repo.GetAudioBookByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestMockRepository_UpdateAudioBook(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	audiobook := &models.AudioBook{
		Title:     "Test Book",
		Author:    "Test Author",
		FilePath:  "/path/to/file.mp3",
		Language:  "en",
		Status:    models.StatusPending,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.CreateAudioBook(ctx, audiobook)
	assert.NoError(t, err)

	// Update the audiobook
	audiobook.Title = "Updated Title"
	err = repo.UpdateAudioBook(ctx, audiobook)
	assert.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetAudioBookByID(ctx, audiobook.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", retrieved.Title)
}

func TestMockRepository_DeleteAudioBook(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	audiobook := &models.AudioBook{
		Title:     "Test Book",
		Author:    "Test Author",
		FilePath:  "/path/to/file.mp3",
		Language:  "en",
		Status:    models.StatusPending,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.CreateAudioBook(ctx, audiobook)
	assert.NoError(t, err)

	// Delete the audiobook
	err = repo.DeleteAudioBook(ctx, audiobook.ID)
	assert.NoError(t, err)

	// Verify it was deleted
	_, err = repo.GetAudioBookByID(ctx, audiobook.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestMockRepository_ListAudioBooks(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Create test audiobooks
	userID := uuid.New()
	for i := 0; i < 5; i++ {
		audiobook := &models.AudioBook{
			Title:     fmt.Sprintf("Test Book %d", i),
			Author:    "Test Author",
			FilePath:  fmt.Sprintf("/path/to/file%d.mp3", i),
			Language:  "en",
			Status:    models.StatusPending,
			CreatedBy: userID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.CreateAudioBook(ctx, audiobook)
		assert.NoError(t, err)
	}

	// Test listing with pagination
	audiobooks, total, err := repo.ListAudioBooks(ctx, 3, 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, audiobooks, 3)

	// Test second page
	audiobooks, total, err = repo.ListAudioBooks(ctx, 3, 3, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, audiobooks, 2)
}

func TestMockRepository_GetAudioBookStats(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	userID := uuid.New()

	// Create audiobooks with different statuses
	audiobooks := []*models.AudioBook{
		{Title: "Book 1", Author: "Author 1", FilePath: "/path1.mp3", Language: "en", Status: models.StatusCompleted, CreatedBy: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "Book 2", Author: "Author 2", FilePath: "/path2.mp3", Language: "en", Status: models.StatusPending, CreatedBy: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "Book 3", Author: "Author 3", FilePath: "/path3.mp3", Language: "en", Status: models.StatusProcessing, CreatedBy: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "Book 4", Author: "Author 4", FilePath: "/path4.mp3", Language: "en", Status: models.StatusFailed, CreatedBy: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, audiobook := range audiobooks {
		err := repo.CreateAudioBook(ctx, audiobook)
		assert.NoError(t, err)
	}

	stats, err := repo.GetAudioBookStats(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 4, stats.TotalAudioBooks)
	assert.Equal(t, 1, stats.CompletedAudioBooks)
	assert.Equal(t, 1, stats.PendingAudioBooks)
	assert.Equal(t, 1, stats.ProcessingAudioBooks)
	assert.Equal(t, 1, stats.FailedAudioBooks)
}
