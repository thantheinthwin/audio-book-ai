package models

import (
	"time"

	"github.com/google/uuid"
)

// Transcript represents a transcript record
type Transcript struct {
	ID                    uuid.UUID `json:"id"`
	AudiobookID           uuid.UUID `json:"audiobook_id"`
	Content               string    `json:"content"`
	Segments              []Segment `json:"segments"`
	Language              string    `json:"language"`
	ConfidenceScore       float64   `json:"confidence_score"`
	ProcessingTimeSeconds int       `json:"processing_time_seconds"`
	CreatedAt             time.Time `json:"created_at"`
}

// Segment represents a transcript segment
type Segment struct {
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Speaker    int     `json:"speaker"`
}

// RevAIJob represents a Rev.ai job submission
type RevAIJob struct {
	MediaURL    string `json:"media_url,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}

// RevAIJobResponse represents the response from Rev.ai job creation
type RevAIJobResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	CreatedOn   string `json:"created_on"`
	Name        string `json:"name"`
	Metadata    string `json:"metadata"`
	CallbackURL string `json:"callback_url"`
}

// RevAITranscript represents the transcript response from Rev.ai
type RevAITranscript struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	CreatedOn   string `json:"created_on"`
	CompletedOn string `json:"completed_on"`
	Monologues  []struct {
		Speaker int `json:"speaker"`
		Elements []struct {
			Type       string  `json:"type"`
			Value      string  `json:"value"`
			StartTs    float64 `json:"start_ts"`
			EndTs      float64 `json:"end_ts"`
			Confidence float64 `json:"confidence"`
		} `json:"elements"`
	} `json:"monologues"`
}
