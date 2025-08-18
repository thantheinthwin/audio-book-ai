package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"audio-book-ai/worker/models"
)

// GeminiService handles Gemini API interactions
type GeminiService struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewGeminiService creates a new Gemini service
func NewGeminiService(apiKey, baseURL, model string) *GeminiService {
	return &GeminiService{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateSummary generates a summary using Gemini API
func (g *GeminiService) GenerateSummary(text string) (string, error) {
	prompt := fmt.Sprintf(`Please provide a concise summary of the following audiobook transcript. Focus on the main themes, key events, and important characters. Keep the summary to 2-3 paragraphs maximum.

Transcript:
%s

Summary:`, text)

	return g.generateText(prompt, 0.3, 1000)
}

// GenerateTags generates tags using Gemini API
func (g *GeminiService) GenerateTags(text string) ([]string, error) {
	prompt := fmt.Sprintf(`Please analyze the following audiobook transcript and generate relevant tags. Focus on:
- Genre (fiction, non-fiction, mystery, romance, etc.)
- Themes (love, betrayal, adventure, etc.)
- Setting (time period, location)
- Target audience (young adult, adult, children, etc.)
- Content warnings if applicable

Return only the tags as a comma-separated list, no explanations.

Transcript:
%s

Tags:`, text)

	response, err := g.generateText(prompt, 0.2, 200)
	if err != nil {
		return nil, err
	}

	// Parse comma-separated tags
	tags := strings.Split(response, ",")
	var cleanTags []string
	for _, tag := range tags {
		cleanTag := strings.TrimSpace(tag)
		if cleanTag != "" {
			cleanTags = append(cleanTags, cleanTag)
		}
	}

	return cleanTags, nil
}

// GenerateEmbedding generates embeddings using Gemini API
func (g *GeminiService) GenerateEmbedding(text string) ([]float64, error) {
	// For now, we'll use a simplified approach
	// In a real implementation, you might want to use a dedicated embedding model
	prompt := fmt.Sprintf(`Please analyze the following text and provide a semantic representation. Focus on the key concepts and themes.

Text:
%s

Analysis:`, text)

	_, err := g.generateText(prompt, 0.1, 500)
	if err != nil {
		return nil, err
	}

	// For now, return a placeholder embedding
	// In a real implementation, you would use a dedicated embedding API
	return []float64{0.1, 0.2, 0.3, 0.4, 0.5}, nil
}

// generateText generates text using Gemini API
func (g *GeminiService) generateText(prompt string, temperature float64, maxTokens int) (string, error) {
	// Truncate text if it's too long for Gemini's limits
	if len(prompt) > 30000 {
		prompt = prompt[:30000] + "..."
	}

	request := models.GeminiRequest{
		Contents: []models.GeminiContent{
			{
				Parts: []models.GeminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: &models.GeminiGenerationConfig{
			Temperature:     temperature,
			MaxOutputTokens: maxTokens,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", g.baseURL, g.model, g.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API error: %d - %s", resp.StatusCode, string(body))
	}

	var response models.GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	// Extract text from the first candidate
	var textParts []string
	for _, part := range response.Candidates[0].Content.Parts {
		textParts = append(textParts, part.Text)
	}

	return strings.Join(textParts, ""), nil
}
