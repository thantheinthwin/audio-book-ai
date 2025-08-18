package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Transcript represents a transcript
type Transcript struct {
	Content string `json:"content"`
}

// AIOutput represents AI processing output
type AIOutput struct {
	ID          uuid.UUID       `json:"id"`
	AudiobookID uuid.UUID       `json:"audiobook_id"`
	OutputType  string          `json:"output_type"`
	Content     json.RawMessage `json:"content"`
	ModelUsed   string          `json:"model_used"`
	CreatedAt   time.Time       `json:"created_at"`
}
