package database

import (
	"context"
	"errors"
	"sync"
	"time"

	"audio-book-ai/api/models"

	"github.com/google/uuid"
)

// Common errors
var ErrNotFound = errors.New("not found")

// MockRepository implements the Repository interface with in-memory storage
type MockRepository struct {
	mu sync.RWMutex

	// In-memory storage
	uploads     map[uuid.UUID]*models.Upload
	uploadFiles map[uuid.UUID]*models.UploadFile
	audiobooks  map[uuid.UUID]*models.AudioBook
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		uploads:     make(map[uuid.UUID]*models.Upload),
		uploadFiles: make(map[uuid.UUID]*models.UploadFile),
		audiobooks:  make(map[uuid.UUID]*models.AudioBook),
	}
}

// Upload operations
func (m *MockRepository) CreateUpload(ctx context.Context, upload *models.Upload) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	upload.ID = uuid.New()
	upload.CreatedAt = time.Now()
	upload.UpdatedAt = time.Now()
	
	m.uploads[upload.ID] = upload
	return nil
}

func (m *MockRepository) GetUploadByID(ctx context.Context, id uuid.UUID) (*models.Upload, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	upload, exists := m.uploads[id]
	if !exists {
		return nil, ErrNotFound
	}
	return upload, nil
}

func (m *MockRepository) GetUploadsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Upload, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var uploads []models.Upload
	for _, upload := range m.uploads {
		if upload.UserID == userID {
			uploads = append(uploads, *upload)
		}
	}
	
	total := len(uploads)
	
	// Simple pagination
	if offset >= total {
		return []models.Upload{}, total, nil
	}
	
	end := offset + limit
	if end > total {
		end = total
	}
	
	return uploads[offset:end], total, nil
}

func (m *MockRepository) UpdateUpload(ctx context.Context, upload *models.Upload) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.uploads[upload.ID]; !exists {
		return ErrNotFound
	}
	
	upload.UpdatedAt = time.Now()
	m.uploads[upload.ID] = upload
	return nil
}

func (m *MockRepository) DeleteUpload(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.uploads[id]; !exists {
		return ErrNotFound
	}
	
	delete(m.uploads, id)
	return nil
}

// Upload File operations
func (m *MockRepository) CreateUploadFile(ctx context.Context, uploadFile *models.UploadFile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	uploadFile.ID = uuid.New()
	uploadFile.CreatedAt = time.Now()
	
	m.uploadFiles[uploadFile.ID] = uploadFile
	return nil
}

func (m *MockRepository) GetUploadFileByID(ctx context.Context, id uuid.UUID) (*models.UploadFile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	uploadFile, exists := m.uploadFiles[id]
	if !exists {
		return nil, ErrNotFound
	}
	return uploadFile, nil
}

func (m *MockRepository) GetUploadFiles(ctx context.Context, uploadID uuid.UUID) ([]models.UploadFile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var files []models.UploadFile
	for _, file := range m.uploadFiles {
		if file.UploadID == uploadID {
			files = append(files, *file)
		}
	}
	return files, nil
}

func (m *MockRepository) UpdateUploadFile(ctx context.Context, uploadFile *models.UploadFile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.uploadFiles[uploadFile.ID]; !exists {
		return ErrNotFound
	}
	
	m.uploadFiles[uploadFile.ID] = uploadFile
	return nil
}

func (m *MockRepository) DeleteUploadFile(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.uploadFiles[id]; !exists {
		return ErrNotFound
	}
	
	delete(m.uploadFiles, id)
	return nil
}

func (m *MockRepository) DeleteUploadFilesByUploadID(ctx context.Context, uploadID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for id, file := range m.uploadFiles {
		if file.UploadID == uploadID {
			delete(m.uploadFiles, id)
		}
	}
	return nil
}

func (m *MockRepository) GetUploadedSize(ctx context.Context, uploadID uuid.UUID) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var totalSize int64
	for _, file := range m.uploadFiles {
		if file.UploadID == uploadID {
			totalSize += file.FileSize
		}
	}
	return totalSize, nil
}

// AudioBook operations (stubs for now)
func (m *MockRepository) CreateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	audiobook.ID = uuid.New()
	audiobook.CreatedAt = time.Now()
	audiobook.UpdatedAt = time.Now()
	
	m.audiobooks[audiobook.ID] = audiobook
	return nil
}

func (m *MockRepository) GetAudioBookByID(ctx context.Context, id uuid.UUID) (*models.AudioBook, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	audiobook, exists := m.audiobooks[id]
	if !exists {
		return nil, ErrNotFound
	}
	return audiobook, nil
}

func (m *MockRepository) GetAudioBookWithDetails(ctx context.Context, id uuid.UUID) (*models.AudioBookWithDetails, error) {
	audiobook, err := m.GetAudioBookByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	return &models.AudioBookWithDetails{
		AudioBook: *audiobook,
	}, nil
}

func (m *MockRepository) UpdateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.audiobooks[audiobook.ID]; !exists {
		return ErrNotFound
	}
	
	audiobook.UpdatedAt = time.Now()
	m.audiobooks[audiobook.ID] = audiobook
	return nil
}

func (m *MockRepository) DeleteAudioBook(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.audiobooks[id]; !exists {
		return ErrNotFound
	}
	
	delete(m.audiobooks, id)
	return nil
}

func (m *MockRepository) ListAudioBooks(ctx context.Context, limit, offset int, isPublic *bool) ([]models.AudioBook, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var audiobooks []models.AudioBook
	for _, audiobook := range m.audiobooks {
		if isPublic == nil || audiobook.IsPublic == *isPublic {
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

func (m *MockRepository) GetAudioBooksByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var audiobooks []models.AudioBook
	for _, audiobook := range m.audiobooks {
		if audiobook.CreatedBy == userID {
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

func (m *MockRepository) UpdateAudioBookStatus(ctx context.Context, id uuid.UUID, status models.AudioBookStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	audiobook, exists := m.audiobooks[id]
	if !exists {
		return ErrNotFound
	}
	
	audiobook.Status = status
	audiobook.UpdatedAt = time.Now()
	return nil
}

// Stub implementations for other interface methods
func (m *MockRepository) CreateChapter(ctx context.Context, chapter *models.Chapter) error {
	return nil
}

func (m *MockRepository) GetChapterByID(ctx context.Context, id uuid.UUID) (*models.Chapter, error) {
	return nil, ErrNotFound
}

func (m *MockRepository) GetChaptersByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Chapter, error) {
	return []models.Chapter{}, nil
}

func (m *MockRepository) GetFirstChapterByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Chapter, error) {
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateChapter(ctx context.Context, chapter *models.Chapter) error {
	return nil
}

func (m *MockRepository) DeleteChapter(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) DeleteChaptersByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateTranscript(ctx context.Context, transcript *models.Transcript) error {
	return nil
}

func (m *MockRepository) GetTranscriptByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Transcript, error) {
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateTranscript(ctx context.Context, transcript *models.Transcript) error {
	return nil
}

func (m *MockRepository) DeleteTranscript(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateChapterTranscript(ctx context.Context, transcript *models.ChapterTranscript) error {
	return nil
}

func (m *MockRepository) GetChapterTranscriptByChapterID(ctx context.Context, chapterID uuid.UUID) (*models.ChapterTranscript, error) {
	return nil, ErrNotFound
}

func (m *MockRepository) GetChapterTranscriptsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ChapterTranscript, error) {
	return []models.ChapterTranscript{}, nil
}

func (m *MockRepository) UpdateChapterTranscript(ctx context.Context, transcript *models.ChapterTranscript) error {
	return nil
}

func (m *MockRepository) DeleteChapterTranscript(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) DeleteChapterTranscriptsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateAIOutput(ctx context.Context, output *models.AIOutput) error {
	return nil
}

func (m *MockRepository) GetAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AIOutput, error) {
	return []models.AIOutput{}, nil
}

func (m *MockRepository) GetAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.AIOutput, error) {
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateAIOutput(ctx context.Context, output *models.AIOutput) error {
	return nil
}

func (m *MockRepository) DeleteAIOutput(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateChapterAIOutput(ctx context.Context, output *models.ChapterAIOutput) error {
	return nil
}

func (m *MockRepository) GetChapterAIOutputsByChapterID(ctx context.Context, chapterID uuid.UUID) ([]models.ChapterAIOutput, error) {
	return []models.ChapterAIOutput{}, nil
}

func (m *MockRepository) GetChapterAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ChapterAIOutput, error) {
	return []models.ChapterAIOutput{}, nil
}

func (m *MockRepository) GetFirstChapterAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.ChapterAIOutput, error) {
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateChapterAIOutput(ctx context.Context, output *models.ChapterAIOutput) error {
	return nil
}

func (m *MockRepository) DeleteChapterAIOutput(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) DeleteChapterAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	return nil
}

func (m *MockRepository) GetProcessingJobsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ProcessingJob, error) {
	return []models.ProcessingJob{}, nil
}

func (m *MockRepository) GetProcessingJobByID(ctx context.Context, id uuid.UUID) (*models.ProcessingJob, error) {
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	return nil
}

func (m *MockRepository) GetPendingJobs(ctx context.Context, jobType models.JobType, limit int) ([]models.ProcessingJob, error) {
	return []models.ProcessingJob{}, nil
}

func (m *MockRepository) GetJobsByStatus(ctx context.Context, status models.JobStatus, limit int) ([]models.ProcessingJob, error) {
	return []models.ProcessingJob{}, nil
}

func (m *MockRepository) CreateTag(ctx context.Context, tag *models.Tag) error {
	return nil
}

func (m *MockRepository) GetTagByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	return nil, ErrNotFound
}

func (m *MockRepository) GetTagByName(ctx context.Context, name string) (*models.Tag, error) {
	return nil, ErrNotFound
}

func (m *MockRepository) GetTagsByCategory(ctx context.Context, category string) ([]models.Tag, error) {
	return []models.Tag{}, nil
}

func (m *MockRepository) UpdateTag(ctx context.Context, tag *models.Tag) error {
	return nil
}

func (m *MockRepository) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) ListTags(ctx context.Context, limit, offset int) ([]models.Tag, int, error) {
	return []models.Tag{}, 0, nil
}

func (m *MockRepository) CreateAudioBookTag(ctx context.Context, audiobookTag *models.AudioBookTag) error {
	return nil
}

func (m *MockRepository) GetTagsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Tag, error) {
	return []models.Tag{}, nil
}

func (m *MockRepository) GetAudioBooksByTagID(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error) {
	return []models.AudioBook{}, 0, nil
}

func (m *MockRepository) DeleteAudioBookTag(ctx context.Context, audiobookID, tagID uuid.UUID) error {
	return nil
}

func (m *MockRepository) DeleteAllAudioBookTags(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (m *MockRepository) CreateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error {
	return nil
}

func (m *MockRepository) GetEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AudioBookEmbedding, error) {
	return []models.AudioBookEmbedding{}, nil
}

func (m *MockRepository) GetEmbeddingByType(ctx context.Context, audiobookID uuid.UUID, embeddingType models.EmbeddingType) (*models.AudioBookEmbedding, error) {
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error {
	return nil
}

func (m *MockRepository) DeleteAudioBookEmbedding(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) DeleteEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	return nil
}

func (m *MockRepository) SearchAudioBooks(ctx context.Context, query string, limit, offset int, language *string, isPublic *bool) ([]models.AudioBook, int, error) {
	return []models.AudioBook{}, 0, nil
}

func (m *MockRepository) SearchAudioBooksByVector(ctx context.Context, embedding []float64, embeddingType models.EmbeddingType, limit, offset int) ([]models.AudioBook, []float64, error) {
	return []models.AudioBook{}, []float64{}, nil
}

func (m *MockRepository) SearchAudioBooksByTags(ctx context.Context, tagNames []string, limit, offset int) ([]models.AudioBook, int, error) {
	return []models.AudioBook{}, 0, nil
}

func (m *MockRepository) GetAudioBookStats(ctx context.Context) (*AudioBookStats, error) {
	return &AudioBookStats{}, nil
}

func (m *MockRepository) GetUserAudioBookStats(ctx context.Context, userID uuid.UUID) (*UserAudioBookStats, error) {
	return &UserAudioBookStats{UserID: userID}, nil
}

func (m *MockRepository) CleanupOrphanedData(ctx context.Context) error {
	return nil
}
