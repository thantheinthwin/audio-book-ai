package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"audio-book-ai/worker/models"
)

// GeminiService handles Gemini API interactions
type GeminiService struct {
	apiKey    string
	baseURL   string
	model     string
	client    *http.Client
	dbService *DatabaseService
}

// NewGeminiService creates a new Gemini service
func NewGeminiService(apiKey, baseURL, model string, dbService *DatabaseService) *GeminiService {
	return &GeminiService{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		dbService: dbService,
	}
}

// GenerateSummaryAndTags generates both summary and tags in a single API call
func (g *GeminiService) GenerateSummaryAndTags(text string) (*models.SummaryAndTags, error) {
	// Get all available tags from the database
	availableTags, err := g.dbService.GetAllTags()
	if err != nil {
		return nil, fmt.Errorf("failed to get available tags: %v", err)
	}

	// Create a list of available tags for the prompt
	availableTagsList := strings.Join(availableTags, ", ")

	prompt := fmt.Sprintf(`Please analyze the following audiobook transcript and provide both a summary and relevant tags.

	Requirements:
	1. Summary: Provide a concise summary (2-3 paragraphs) focusing on main themes, key events, and important characters.
	2. Tags: Choose ONLY from the following available tags: %s

	IMPORTANT: 
	- You must ONLY use tags from the provided list. Do not create new tags.
	- Respond with ONLY the JSON object, no markdown formatting, no code blocks, no additional text.

	Please respond with ONLY this JSON format:
	{
	"summary": "Your summary here...",
	"tags": ["tag1", "tag2", "tag3", ...]
	}

	Available tags: %s

	Transcript:
	%s

	Response:`, availableTagsList, availableTagsList, text)

	response, err := g.generateText(prompt, 0.3, 1000)
	if err != nil {
		return nil, err
	}

	// Try to parse the response as JSON
	var summaryAndTags models.SummaryAndTags

	// First, try to parse the response directly as JSON
	if err := json.Unmarshal([]byte(response), &summaryAndTags); err == nil {
		// Successfully parsed as JSON
		log.Printf("Successfully parsed JSON response directly")
	} else {
		log.Printf("Direct JSON parsing failed, attempting to extract from markdown: %v", err)
		// If direct JSON parsing fails, try to extract JSON from markdown code blocks
		cleanedResponse := g.extractJSONFromMarkdown(response)
		if cleanedResponse != "" {
			// log.Printf("Extracted JSON from markdown: %s", cleanedResponse)
			if err := json.Unmarshal([]byte(cleanedResponse), &summaryAndTags); err == nil {
				// Successfully parsed JSON from markdown
				log.Printf("Successfully parsed JSON from markdown")
			} else {
				log.Printf("JSON parsing from markdown failed: %v", err)
				// If JSON parsing still fails, try to extract summary and tags manually
				return g.extractSummaryAndTagsFromText(response)
			}
		} else {
			log.Printf("No JSON found in markdown, falling back to text extraction")
			// If no JSON found in markdown, try to extract summary and tags manually
			return g.extractSummaryAndTagsFromText(response)
		}
	}

	// Filter tags to ensure only valid tags from the database are used
	summaryAndTags.Tags = g.filterValidTags(summaryAndTags.Tags, availableTags)

	fmt.Println("Summary and tags:", summaryAndTags)

	return &summaryAndTags, nil
}

// extractJSONFromMarkdown extracts JSON content from markdown code blocks
func (g *GeminiService) extractJSONFromMarkdown(text string) string {
	// Look for JSON code blocks (```json ... ```)
	startMarker := "```json"
	endMarker := "```"

	startIndex := strings.Index(strings.ToLower(text), startMarker)
	if startIndex == -1 {
		// Also try without the language specifier
		startMarker = "```"
		startIndex = strings.Index(text, startMarker)
		if startIndex == -1 {
			return ""
		}
	}

	// Find the end of the start marker
	startContent := startIndex + len(startMarker)

	// Find the end marker
	endIndex := strings.Index(text[startContent:], endMarker)
	if endIndex == -1 {
		return ""
	}

	// Extract the content between the markers
	jsonContent := text[startContent : startContent+endIndex]

	// Clean up the content
	jsonContent = strings.TrimSpace(jsonContent)

	return jsonContent
}

// filterValidTags filters the provided tags to only include those that exist in the available tags list
func (g *GeminiService) filterValidTags(providedTags []string, availableTags []string) []string {
	// Create a map for O(1) lookup
	availableTagsMap := make(map[string]bool)
	for _, tag := range availableTags {
		availableTagsMap[strings.ToLower(strings.TrimSpace(tag))] = true
	}

	var validTags []string
	for _, tag := range providedTags {
		tag = strings.TrimSpace(tag)
		if availableTagsMap[strings.ToLower(tag)] {
			validTags = append(validTags, tag)
		}
	}

	return validTags
}

// extractSummaryAndTagsFromText extracts summary and tags from plain text response
func (g *GeminiService) extractSummaryAndTagsFromText(text string) (*models.SummaryAndTags, error) {
	// Get all available tags from the database
	availableTags, err := g.dbService.GetAllTags()
	if err != nil {
		return nil, fmt.Errorf("failed to get available tags: %v", err)
	}
	// Simple extraction logic - look for common patterns
	lines := strings.Split(text, "\n")
	var summaryLines []string
	var tags []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for tag indicators
		if strings.Contains(strings.ToLower(line), "tags:") ||
			strings.Contains(strings.ToLower(line), "tag:") ||
			strings.HasPrefix(line, "-") ||
			strings.HasPrefix(line, "*") {
			// Extract tags from this line
			tagLine := strings.TrimPrefix(line, "Tags:")
			tagLine = strings.TrimPrefix(tagLine, "Tag:")
			tagLine = strings.TrimSpace(tagLine)

			// Split by commas or other separators
			potentialTags := strings.Split(tagLine, ",")
			for _, tag := range potentialTags {
				tag = strings.TrimSpace(tag)
				tag = strings.TrimPrefix(tag, "-")
				tag = strings.TrimPrefix(tag, "*")
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tags = append(tags, tag)
				}
			}
		} else {
			// Assume it's part of the summary
			summaryLines = append(summaryLines, line)
		}
	}

	summary := strings.Join(summaryLines, "\n")
	if summary == "" {
		summary = "Summary could not be extracted from response."
	}

	// Filter tags to ensure only valid tags from the database are used
	filteredTags := g.filterValidTags(tags, availableTags)

	return &models.SummaryAndTags{
		Summary: summary,
		Tags:    filteredTags,
	}, nil
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
