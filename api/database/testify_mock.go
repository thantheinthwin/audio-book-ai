package database

import (
	"audio-book-ai/api/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// Common errors
var ErrNotFound = errors.New("not found")

// TestifyMockRepository is a testify/mock based implementation of the Repository interface
type TestifyMockRepository struct {
	mock.Mock
}

// AudioBook operations
func (m *TestifyMockRepository) CreateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
	args := m.Called(ctx, audiobook)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetAudioBookByID(ctx context.Context, id uuid.UUID) (*models.AudioBook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AudioBook), args.Error(1)
}

func (m *TestifyMockRepository) GetAudioBookWithDetails(ctx context.Context, id uuid.UUID) (*models.AudioBookWithDetails, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AudioBookWithDetails), args.Error(1)
}

func (m *TestifyMockRepository) UpdateAudioBook(ctx context.Context, audiobook *models.AudioBook) error {
	args := m.Called(ctx, audiobook)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteAudioBook(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TestifyMockRepository) ListAudioBooks(ctx context.Context, limit, offset int, isPublic *bool) ([]models.AudioBook, int, error) {
	args := m.Called(ctx, limit, offset, isPublic)
	return args.Get(0).([]models.AudioBook), args.Int(1), args.Error(2)
}

func (m *TestifyMockRepository) GetAudioBooksByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.AudioBook), args.Int(1), args.Error(2)
}

func (m *TestifyMockRepository) UpdateAudioBookStatus(ctx context.Context, id uuid.UUID, status models.AudioBookStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *TestifyMockRepository) CheckAndUpdateAudioBookStatus(ctx context.Context, audiobookID uuid.UUID) error {
	args := m.Called(ctx, audiobookID)
	return args.Error(0)
}

// Chapter operations
func (m *TestifyMockRepository) CreateChapter(ctx context.Context, chapter *models.Chapter) error {
	args := m.Called(ctx, chapter)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetChapterByID(ctx context.Context, id uuid.UUID) (*models.Chapter, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chapter), args.Error(1)
}

func (m *TestifyMockRepository) GetChaptersByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Chapter, error) {
	args := m.Called(ctx, audiobookID)
	return args.Get(0).([]models.Chapter), args.Error(1)
}

func (m *TestifyMockRepository) GetFirstChapterByAudioBookID(ctx context.Context, audiobookID uuid.UUID) (*models.Chapter, error) {
	args := m.Called(ctx, audiobookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chapter), args.Error(1)
}

func (m *TestifyMockRepository) UpdateChapter(ctx context.Context, chapter *models.Chapter) error {
	args := m.Called(ctx, chapter)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteChapter(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteChaptersByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	args := m.Called(ctx, audiobookID)
	return args.Error(0)
}

// Chapter Transcript operations
func (m *TestifyMockRepository) CreateChapterTranscript(ctx context.Context, transcript *models.ChapterTranscript) error {
	args := m.Called(ctx, transcript)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetChapterTranscriptByChapterID(ctx context.Context, chapterID uuid.UUID) (*models.ChapterTranscript, error) {
	args := m.Called(ctx, chapterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChapterTranscript), args.Error(1)
}

func (m *TestifyMockRepository) GetChapterTranscriptsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ChapterTranscript, error) {
	args := m.Called(ctx, audiobookID)
	return args.Get(0).([]models.ChapterTranscript), args.Error(1)
}

func (m *TestifyMockRepository) UpdateChapterTranscript(ctx context.Context, transcript *models.ChapterTranscript) error {
	args := m.Called(ctx, transcript)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteChapterTranscript(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteChapterTranscriptsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	args := m.Called(ctx, audiobookID)
	return args.Error(0)
}

// AI Output operations
func (m *TestifyMockRepository) CreateAIOutput(ctx context.Context, output *models.AIOutput) error {
	args := m.Called(ctx, output)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AIOutput, error) {
	args := m.Called(ctx, audiobookID)
	return args.Get(0).([]models.AIOutput), args.Error(1)
}

func (m *TestifyMockRepository) GetAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.AIOutput, error) {
	args := m.Called(ctx, audiobookID, outputType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AIOutput), args.Error(1)
}

func (m *TestifyMockRepository) UpdateAIOutput(ctx context.Context, output *models.AIOutput) error {
	args := m.Called(ctx, output)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteAIOutput(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Chapter AI Output operations
func (m *TestifyMockRepository) CreateChapterAIOutput(ctx context.Context, output *models.ChapterAIOutput) error {
	args := m.Called(ctx, output)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetChapterAIOutputsByChapterID(ctx context.Context, chapterID uuid.UUID) ([]models.ChapterAIOutput, error) {
	args := m.Called(ctx, chapterID)
	return args.Get(0).([]models.ChapterAIOutput), args.Error(1)
}

func (m *TestifyMockRepository) GetChapterAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ChapterAIOutput, error) {
	args := m.Called(ctx, audiobookID)
	return args.Get(0).([]models.ChapterAIOutput), args.Error(1)
}

func (m *TestifyMockRepository) GetFirstChapterAIOutputByType(ctx context.Context, audiobookID uuid.UUID, outputType models.OutputType) (*models.ChapterAIOutput, error) {
	args := m.Called(ctx, audiobookID, outputType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChapterAIOutput), args.Error(1)
}

func (m *TestifyMockRepository) UpdateChapterAIOutput(ctx context.Context, output *models.ChapterAIOutput) error {
	args := m.Called(ctx, output)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteChapterAIOutput(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteChapterAIOutputsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	args := m.Called(ctx, audiobookID)
	return args.Error(0)
}

// Processing Job operations
func (m *TestifyMockRepository) CreateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetProcessingJobsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.ProcessingJob, error) {
	args := m.Called(ctx, audiobookID)
	return args.Get(0).([]models.ProcessingJob), args.Error(1)
}

func (m *TestifyMockRepository) GetProcessingJobByID(ctx context.Context, id uuid.UUID) (*models.ProcessingJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProcessingJob), args.Error(1)
}

func (m *TestifyMockRepository) UpdateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetPendingJobs(ctx context.Context, jobType models.JobType, limit int) ([]models.ProcessingJob, error) {
	args := m.Called(ctx, jobType, limit)
	return args.Get(0).([]models.ProcessingJob), args.Error(1)
}

func (m *TestifyMockRepository) GetJobsByStatus(ctx context.Context, status models.JobStatus, limit int) ([]models.ProcessingJob, error) {
	args := m.Called(ctx, status, limit)
	return args.Get(0).([]models.ProcessingJob), args.Error(1)
}

func (m *TestifyMockRepository) IncrementRetryCount(ctx context.Context, jobID uuid.UUID) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func (m *TestifyMockRepository) ResetRetryCount(ctx context.Context, jobID uuid.UUID) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

// Tag operations
func (m *TestifyMockRepository) CreateTag(ctx context.Context, tag *models.Tag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetTagByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tag), args.Error(1)
}

func (m *TestifyMockRepository) GetTagByName(ctx context.Context, name string) (*models.Tag, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tag), args.Error(1)
}

func (m *TestifyMockRepository) GetTagsByCategory(ctx context.Context, category string) ([]models.Tag, error) {
	args := m.Called(ctx, category)
	return args.Get(0).([]models.Tag), args.Error(1)
}

func (m *TestifyMockRepository) UpdateTag(ctx context.Context, tag *models.Tag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteTag(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TestifyMockRepository) ListTags(ctx context.Context, limit, offset int) ([]models.Tag, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]models.Tag), args.Int(1), args.Error(2)
}

// AudioBook Tag operations
func (m *TestifyMockRepository) CreateAudioBookTag(ctx context.Context, audiobookTag *models.AudioBookTag) error {
	args := m.Called(ctx, audiobookTag)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetTagsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.Tag, error) {
	args := m.Called(ctx, audiobookID)
	return args.Get(0).([]models.Tag), args.Error(1)
}

func (m *TestifyMockRepository) GetAudioBooksByTagID(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]models.AudioBook, int, error) {
	args := m.Called(ctx, tagID, limit, offset)
	return args.Get(0).([]models.AudioBook), args.Int(1), args.Error(2)
}

func (m *TestifyMockRepository) DeleteAudioBookTag(ctx context.Context, audiobookID, tagID uuid.UUID) error {
	args := m.Called(ctx, audiobookID, tagID)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteAllAudioBookTags(ctx context.Context, audiobookID uuid.UUID) error {
	args := m.Called(ctx, audiobookID)
	return args.Error(0)
}

// AudioBook Embedding operations
func (m *TestifyMockRepository) CreateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error {
	args := m.Called(ctx, embedding)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) ([]models.AudioBookEmbedding, error) {
	args := m.Called(ctx, audiobookID)
	return args.Get(0).([]models.AudioBookEmbedding), args.Error(1)
}

func (m *TestifyMockRepository) GetEmbeddingByType(ctx context.Context, audiobookID uuid.UUID, embeddingType models.EmbeddingType) (*models.AudioBookEmbedding, error) {
	args := m.Called(ctx, audiobookID, embeddingType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AudioBookEmbedding), args.Error(1)
}

func (m *TestifyMockRepository) UpdateAudioBookEmbedding(ctx context.Context, embedding *models.AudioBookEmbedding) error {
	args := m.Called(ctx, embedding)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteAudioBookEmbedding(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteEmbeddingsByAudioBookID(ctx context.Context, audiobookID uuid.UUID) error {
	args := m.Called(ctx, audiobookID)
	return args.Error(0)
}

// Upload operations
func (m *TestifyMockRepository) CreateUpload(ctx context.Context, upload *models.Upload) error {
	args := m.Called(ctx, upload)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetUploadByID(ctx context.Context, id uuid.UUID) (*models.Upload, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Upload), args.Error(1)
}

func (m *TestifyMockRepository) GetUploadsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Upload, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.Upload), args.Int(1), args.Error(2)
}

func (m *TestifyMockRepository) UpdateUpload(ctx context.Context, upload *models.Upload) error {
	args := m.Called(ctx, upload)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteUpload(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Upload File operations
func (m *TestifyMockRepository) CreateUploadFile(ctx context.Context, uploadFile *models.UploadFile) error {
	args := m.Called(ctx, uploadFile)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetUploadFileByID(ctx context.Context, id uuid.UUID) (*models.UploadFile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UploadFile), args.Error(1)
}

func (m *TestifyMockRepository) GetUploadFiles(ctx context.Context, uploadID uuid.UUID) ([]models.UploadFile, error) {
	args := m.Called(ctx, uploadID)
	return args.Get(0).([]models.UploadFile), args.Error(1)
}

func (m *TestifyMockRepository) UpdateUploadFile(ctx context.Context, uploadFile *models.UploadFile) error {
	args := m.Called(ctx, uploadFile)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteUploadFile(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TestifyMockRepository) DeleteUploadFilesByUploadID(ctx context.Context, uploadID uuid.UUID) error {
	args := m.Called(ctx, uploadID)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetUploadedSize(ctx context.Context, uploadID uuid.UUID) (int64, error) {
	args := m.Called(ctx, uploadID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *TestifyMockRepository) GetFailedUploadFiles(ctx context.Context, uploadID uuid.UUID) ([]models.UploadFile, error) {
	args := m.Called(ctx, uploadID)
	return args.Get(0).([]models.UploadFile), args.Error(1)
}

func (m *TestifyMockRepository) GetRetryingUploadFiles(ctx context.Context, uploadID uuid.UUID) ([]models.UploadFile, error) {
	args := m.Called(ctx, uploadID)
	return args.Get(0).([]models.UploadFile), args.Error(1)
}

func (m *TestifyMockRepository) IncrementUploadFileRetryCount(ctx context.Context, fileID uuid.UUID) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *TestifyMockRepository) ResetUploadFileRetryCount(ctx context.Context, fileID uuid.UUID) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

// Search operations
func (m *TestifyMockRepository) SearchAudioBooks(ctx context.Context, query string, limit, offset int, language *string, isPublic *bool) ([]models.AudioBook, int, error) {
	args := m.Called(ctx, query, limit, offset, language, isPublic)
	return args.Get(0).([]models.AudioBook), args.Int(1), args.Error(2)
}

func (m *TestifyMockRepository) SearchAudioBooksByVector(ctx context.Context, embedding []float64, embeddingType models.EmbeddingType, limit, offset int) ([]models.AudioBook, []float64, error) {
	args := m.Called(ctx, embedding, embeddingType, limit, offset)
	return args.Get(0).([]models.AudioBook), args.Get(1).([]float64), args.Error(2)
}

func (m *TestifyMockRepository) SearchAudioBooksByTags(ctx context.Context, tagNames []string, limit, offset int) ([]models.AudioBook, int, error) {
	args := m.Called(ctx, tagNames, limit, offset)
	return args.Get(0).([]models.AudioBook), args.Int(1), args.Error(2)
}

// Utility operations
func (m *TestifyMockRepository) GetAudioBookStats(ctx context.Context) (*AudioBookStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AudioBookStats), args.Error(1)
}

func (m *TestifyMockRepository) GetUserAudioBookStats(ctx context.Context, userID uuid.UUID) (*UserAudioBookStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserAudioBookStats), args.Error(1)
}

func (m *TestifyMockRepository) CleanupOrphanedData(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Cart operations
func (m *TestifyMockRepository) AddToCart(ctx context.Context, userID, audiobookID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID, audiobookID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *TestifyMockRepository) RemoveFromCart(ctx context.Context, userID, audiobookID uuid.UUID) error {
	args := m.Called(ctx, userID, audiobookID)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetCartItems(ctx context.Context, userID uuid.UUID) ([]models.CartItemWithDetails, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.CartItemWithDetails), args.Error(1)
}

func (m *TestifyMockRepository) IsInCart(ctx context.Context, userID, audiobookID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, audiobookID)
	return args.Bool(0), args.Error(1)
}

func (m *TestifyMockRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Purchased Audiobook operations
func (m *TestifyMockRepository) CreatePurchasedAudioBook(ctx context.Context, purchase *models.PurchasedAudioBook) error {
	args := m.Called(ctx, purchase)
	return args.Error(0)
}

func (m *TestifyMockRepository) GetPurchasedAudioBookByID(ctx context.Context, id uuid.UUID) (*models.PurchasedAudioBook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchasedAudioBook), args.Error(1)
}

func (m *TestifyMockRepository) GetPurchasedAudioBooksByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.PurchasedAudioBookWithDetails, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.PurchasedAudioBookWithDetails), args.Int(1), args.Error(2)
}

func (m *TestifyMockRepository) IsAudioBookPurchased(ctx context.Context, userID, audiobookID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, audiobookID)
	return args.Bool(0), args.Error(1)
}

func (m *TestifyMockRepository) GetPurchaseHistory(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.PurchaseHistoryResponse, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchaseHistoryResponse), args.Error(1)
}

func (m *TestifyMockRepository) GetPurchasedAudioBookByUserAndAudiobook(ctx context.Context, userID, audiobookID uuid.UUID) (*models.PurchasedAudioBook, error) {
	args := m.Called(ctx, userID, audiobookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PurchasedAudioBook), args.Error(1)
}
