package models

import (
	"time"

	"github.com/google/uuid"
)

// AIOutput represents an AI processing output
type AIOutput struct {
	ID          uuid.UUID `json:"id"`
	AudiobookID uuid.UUID `json:"audiobook_id"`
	OutputType  string    `json:"output_type"`
	Content     any       `json:"content"`
	ModelUsed   string    `json:"model_used"`
	CreatedAt   time.Time `json:"created_at"`
	ProcessingTimeSeconds int `json:"processing_time_seconds"`
}

// OutputType represents the possible output types
const (
	OutputTypeSummary   = "summary"
	OutputTypeEmbedding = "embedding"
)

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Contents         []GeminiContent         `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent represents content in a Gemini request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of content in a Gemini request
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig represents generation configuration for Gemini
type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	TopK            int     `json:"topK,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

// GeminiResponse represents a response from the Gemini API
type GeminiResponse struct {
	Candidates     []GeminiCandidate     `json:"candidates"`
	PromptFeedback *GeminiPromptFeedback `json:"promptFeedback,omitempty"`
}

// GeminiCandidate represents a candidate response from Gemini
type GeminiCandidate struct {
	Content       GeminiContent        `json:"content"`
	FinishReason  string               `json:"finishReason"`
	Index         int                  `json:"index"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

// GeminiPromptFeedback represents feedback about the prompt
type GeminiPromptFeedback struct {
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

// GeminiSafetyRating represents a safety rating
type GeminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// ChapterTranscript represents a chapter transcript for processing
type ChapterTranscript struct {
	ID                    uuid.UUID `json:"id"`
	ChapterID             uuid.UUID `json:"chapter_id"`
	AudiobookID           uuid.UUID `json:"audiobook_id"`
	Content               string    `json:"content"`
	Segments              []byte    `json:"segments,omitempty"`
	Language              *string   `json:"language,omitempty"`
	ConfidenceScore       *float64  `json:"confidence_score,omitempty"`
	ProcessingTimeSeconds *int      `json:"processing_time_seconds,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}

// SummaryAndTags represents the combined response from Gemini for summary and tags
type SummaryAndTags struct {
	Summary string   `json:"summary"`
	Tags    []string `json:"tags"`
}
