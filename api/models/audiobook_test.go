package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAudioBook_Validation(t *testing.T) {
	tests := []struct {
		name      string
		audiobook AudioBook
		wantErr   bool
	}{
		{
			name: "valid audio book",
			audiobook: AudioBook{
				ID:        uuid.New(),
				Title:     "Test Book",
				Author:    "Test Author",
				Language:  "en",
				Status:    StatusPending,
				CreatedBy: uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing title",
			audiobook: AudioBook{
				ID:        uuid.New(),
				Author:    "Test Author",
				Language:  "en",
				Status:    StatusPending,
				CreatedBy: uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing author",
			audiobook: AudioBook{
				ID:        uuid.New(),
				Title:     "Test Book",
				Language:  "en",
				Status:    StatusPending,
				CreatedBy: uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing file path",
			audiobook: AudioBook{
				ID:        uuid.New(),
				Title:     "Test Book",
				Author:    "Test Author",
				Language:  "en",
				Status:    StatusPending,
				CreatedBy: uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid language code",
			audiobook: AudioBook{
				ID:        uuid.New(),
				Title:     "Test Book",
				Author:    "Test Author",
				Language:  "english",
				Status:    StatusPending,
				CreatedBy: uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing created by",
			audiobook: AudioBook{
				ID:        uuid.New(),
				Title:     "Test Book",
				Author:    "Test Author",
				Language:  "en",
				Status:    StatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.audiobook.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChapter_Validation(t *testing.T) {
	tests := []struct {
		name    string
		chapter Chapter
		wantErr bool
	}{
		{
			name: "valid chapter",
			chapter: Chapter{
				ID:            uuid.New(),
				AudiobookID:   uuid.New(),
				ChapterNumber: 1,
				Title:         "Chapter 1: Introduction",
				CreatedAt:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing audiobook ID",
			chapter: Chapter{
				ID:            uuid.New(),
				ChapterNumber: 1,
				Title:         "Chapter 1: Introduction",
				CreatedAt:     time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid chapter number",
			chapter: Chapter{
				ID:            uuid.New(),
				AudiobookID:   uuid.New(),
				ChapterNumber: 0,
				Title:         "Chapter 1: Introduction",
				CreatedAt:     time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing title",
			chapter: Chapter{
				ID:            uuid.New(),
				AudiobookID:   uuid.New(),
				ChapterNumber: 1,
				CreatedAt:     time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty title",
			chapter: Chapter{
				ID:            uuid.New(),
				AudiobookID:   uuid.New(),
				ChapterNumber: 1,
				Title:         "",
				CreatedAt:     time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.chapter.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChapter_IsFirstChapter(t *testing.T) {
	tests := []struct {
		name           string
		chapterNumber  int
		expectedResult bool
	}{
		{
			name:           "first chapter",
			chapterNumber:  1,
			expectedResult: true,
		},
		{
			name:           "second chapter",
			chapterNumber:  2,
			expectedResult: false,
		},
		{
			name:           "tenth chapter",
			chapterNumber:  10,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chapter := Chapter{ChapterNumber: tt.chapterNumber}
			assert.Equal(t, tt.expectedResult, chapter.IsFirstChapter())
		})
	}
}

func TestChapterTranscript_Validation(t *testing.T) {
	tests := []struct {
		name       string
		transcript ChapterTranscript
		wantErr    bool
	}{
		{
			name: "valid chapter transcript",
			transcript: ChapterTranscript{
				ID:          uuid.New(),
				ChapterID:   uuid.New(),
				AudiobookID: uuid.New(),
				Content:     "This is the chapter transcript content.",
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing chapter ID",
			transcript: ChapterTranscript{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				Content:     "This is the chapter transcript content.",
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing audiobook ID",
			transcript: ChapterTranscript{
				ID:        uuid.New(),
				ChapterID: uuid.New(),
				Content:   "This is the chapter transcript content.",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing content",
			transcript: ChapterTranscript{
				ID:          uuid.New(),
				ChapterID:   uuid.New(),
				AudiobookID: uuid.New(),
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transcript.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChapterAIOutput_Validation(t *testing.T) {
	validContent := json.RawMessage(`{"summary": "This is a chapter summary"}`)

	tests := []struct {
		name     string
		aiOutput ChapterAIOutput
		wantErr  bool
	}{
		{
			name: "valid chapter AI output",
			aiOutput: ChapterAIOutput{
				ID:          uuid.New(),
				ChapterID:   uuid.New(),
				AudiobookID: uuid.New(),
				OutputType:  OutputTypeSummary,
				Content:     validContent,
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing chapter ID",
			aiOutput: ChapterAIOutput{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				OutputType:  OutputTypeSummary,
				Content:     validContent,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing audiobook ID",
			aiOutput: ChapterAIOutput{
				ID:         uuid.New(),
				ChapterID:  uuid.New(),
				OutputType: OutputTypeSummary,
				Content:    validContent,
				CreatedAt:  time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing output type",
			aiOutput: ChapterAIOutput{
				ID:          uuid.New(),
				ChapterID:   uuid.New(),
				AudiobookID: uuid.New(),
				Content:     validContent,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing content",
			aiOutput: ChapterAIOutput{
				ID:          uuid.New(),
				ChapterID:   uuid.New(),
				AudiobookID: uuid.New(),
				OutputType:  OutputTypeSummary,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.aiOutput.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAudioBook_StatusMethods(t *testing.T) {
	tests := []struct {
		name     string
		status   AudioBookStatus
		expected map[string]bool
	}{
		{
			name:   "pending status",
			status: StatusPending,
			expected: map[string]bool{
				"isCompleted":  false,
				"isFailed":     false,
				"isProcessing": false,
				"isPending":    true,
			},
		},
		{
			name:   "processing status",
			status: StatusProcessing,
			expected: map[string]bool{
				"isCompleted":  false,
				"isFailed":     false,
				"isProcessing": true,
				"isPending":    false,
			},
		},
		{
			name:   "completed status",
			status: StatusCompleted,
			expected: map[string]bool{
				"isCompleted":  true,
				"isFailed":     false,
				"isProcessing": false,
				"isPending":    false,
			},
		},
		{
			name:   "failed status",
			status: StatusFailed,
			expected: map[string]bool{
				"isCompleted":  false,
				"isFailed":     true,
				"isProcessing": false,
				"isPending":    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audiobook := AudioBook{Status: tt.status}

			assert.Equal(t, tt.expected["isCompleted"], audiobook.IsCompleted())
			assert.Equal(t, tt.expected["isFailed"], audiobook.IsFailed())
			assert.Equal(t, tt.expected["isProcessing"], audiobook.IsProcessing())
			assert.Equal(t, tt.expected["isPending"], audiobook.IsPending())
		})
	}
}

func TestAudioBook_GetDurationFormatted(t *testing.T) {
	tests := []struct {
		name     string
		seconds  *int
		expected string
	}{
		{
			name:     "nil duration",
			seconds:  nil,
			expected: "Unknown",
		},
		{
			name:     "30 seconds",
			seconds:  intPtr(30),
			expected: "0m 30s",
		},
		{
			name:     "90 seconds",
			seconds:  intPtr(90),
			expected: "1m 30s",
		},
		{
			name:     "3661 seconds",
			seconds:  intPtr(3661),
			expected: "1h 1m 1s",
		},
		{
			name:     "7200 seconds",
			seconds:  intPtr(7200),
			expected: "2h 0m 0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audiobook := AudioBook{DurationSeconds: tt.seconds}
			assert.Equal(t, tt.expected, audiobook.GetDurationFormatted())
		})
	}
}

func TestTranscript_Validation(t *testing.T) {
	tests := []struct {
		name       string
		transcript Transcript
		wantErr    bool
	}{
		{
			name: "valid transcript",
			transcript: Transcript{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				Content:     "This is the transcript content.",
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing audiobook ID",
			transcript: Transcript{
				ID:        uuid.New(),
				Content:   "This is the transcript content.",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing content",
			transcript: Transcript{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transcript.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAIOutput_Validation(t *testing.T) {
	validContent := json.RawMessage(`{"summary": "This is a summary"}`)

	tests := []struct {
		name     string
		aiOutput AIOutput
		wantErr  bool
	}{
		{
			name: "valid AI output",
			aiOutput: AIOutput{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				OutputType:  OutputTypeSummary,
				Content:     validContent,
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing audiobook ID",
			aiOutput: AIOutput{
				ID:         uuid.New(),
				OutputType: OutputTypeSummary,
				Content:    validContent,
				CreatedAt:  time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing output type",
			aiOutput: AIOutput{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				Content:     validContent,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing content",
			aiOutput: AIOutput{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				OutputType:  OutputTypeSummary,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.aiOutput.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessingJob_Validation(t *testing.T) {
	tests := []struct {
		name    string
		job     ProcessingJob
		wantErr bool
	}{
		{
			name: "valid processing job",
			job: ProcessingJob{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				JobType:     JobTypeTranscribe,
				Status:      JobStatusPending,
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing audiobook ID",
			job: ProcessingJob{
				ID:        uuid.New(),
				JobType:   JobTypeTranscribe,
				Status:    JobStatusPending,
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing job type",
			job: ProcessingJob{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				Status:      JobStatusPending,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing status",
			job: ProcessingJob{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				JobType:     JobTypeTranscribe,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTag_Validation(t *testing.T) {
	tests := []struct {
		name    string
		tag     Tag
		wantErr bool
	}{
		{
			name: "valid tag",
			tag: Tag{
				ID:        uuid.New(),
				Name:      "Fiction",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			tag: Tag{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty name",
			tag: Tag{
				ID:        uuid.New(),
				Name:      "",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tag.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAudioBookEmbedding_Validation(t *testing.T) {
	validEmbedding := []float64{0.1, 0.2, 0.3}

	tests := []struct {
		name      string
		embedding AudioBookEmbedding
		wantErr   bool
	}{
		{
			name: "valid embedding",
			embedding: AudioBookEmbedding{
				ID:            uuid.New(),
				AudiobookID:   uuid.New(),
				Embedding:     validEmbedding,
				EmbeddingType: EmbeddingTypeTitle,
				CreatedAt:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing audiobook ID",
			embedding: AudioBookEmbedding{
				ID:            uuid.New(),
				Embedding:     validEmbedding,
				EmbeddingType: EmbeddingTypeTitle,
				CreatedAt:     time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing embedding",
			embedding: AudioBookEmbedding{
				ID:            uuid.New(),
				AudiobookID:   uuid.New(),
				EmbeddingType: EmbeddingTypeTitle,
				CreatedAt:     time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing embedding type",
			embedding: AudioBookEmbedding{
				ID:          uuid.New(),
				AudiobookID: uuid.New(),
				Embedding:   validEmbedding,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.embedding.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateAudioBookRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateAudioBookRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: CreateAudioBookRequest{
				Title:    "Test Book",
				Author:   "Test Author",
				Language: "en",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			request: CreateAudioBookRequest{
				Author:   "Test Author",
				Language: "en",
			},
			wantErr: true,
		},
		{
			name: "missing author",
			request: CreateAudioBookRequest{
				Title:    "Test Book",
				Language: "en",
			},
			wantErr: true,
		},
		{
			name: "missing language",
			request: CreateAudioBookRequest{
				Title:  "Test Book",
				Author: "Test Author",
			},
			wantErr: true,
		},
		{
			name: "invalid language",
			request: CreateAudioBookRequest{
				Title:    "Test Book",
				Author:   "Test Author",
				Language: "english",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request SearchRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: SearchRequest{
				Query:  "test query",
				Limit:  10,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "missing query",
			request: SearchRequest{
				Limit:  10,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "empty query",
			request: SearchRequest{
				Query:  "",
				Limit:  10,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "limit too low",
			request: SearchRequest{
				Query:  "test query",
				Limit:  0,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "limit too high",
			request: SearchRequest{
				Query:  "test query",
				Limit:  101,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "negative offset",
			request: SearchRequest{
				Query:  "test query",
				Limit:  10,
				Offset: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAudioBookTag_Validation(t *testing.T) {
	tests := []struct {
		name    string
		tag     AudioBookTag
		wantErr bool
	}{
		{
			name: "valid audio book tag",
			tag: AudioBookTag{
				AudiobookID:   uuid.New(),
				TagID:         uuid.New(),
				IsAIGenerated: false,
				CreatedAt:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing audiobook ID",
			tag: AudioBookTag{
				TagID:         uuid.New(),
				IsAIGenerated: false,
				CreatedAt:     time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing tag ID",
			tag: AudioBookTag{
				AudiobookID:   uuid.New(),
				IsAIGenerated: false,
				CreatedAt:     time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tag.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateAudioBookRequest_Validation(t *testing.T) {
	validTitle := "Updated Title"
	validAuthor := "Updated Author"
	validLanguage := "es"

	tests := []struct {
		name    string
		request UpdateAudioBookRequest
		wantErr bool
	}{
		{
			name: "valid request with title",
			request: UpdateAudioBookRequest{
				Title: &validTitle,
			},
			wantErr: false,
		},
		{
			name: "valid request with author",
			request: UpdateAudioBookRequest{
				Author: &validAuthor,
			},
			wantErr: false,
		},
		{
			name: "valid request with language",
			request: UpdateAudioBookRequest{
				Language: &validLanguage,
			},
			wantErr: false,
		},
		{
			name: "empty title",
			request: UpdateAudioBookRequest{
				Title: stringPtr(""),
			},
			wantErr: true,
		},
		{
			name: "empty author",
			request: UpdateAudioBookRequest{
				Author: stringPtr(""),
			},
			wantErr: true,
		},
		{
			name: "invalid language",
			request: UpdateAudioBookRequest{
				Language: stringPtr("english"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
