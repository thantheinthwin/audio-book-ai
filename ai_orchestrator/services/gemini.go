package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"

    "audio-book-ai/ai_orchestrator/models"
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
        client: &http.Client{Timeout: 60 * time.Second},
    }
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

    parts := strings.Split(response, ",")
    var tags []string
    for _, p := range parts {
        t := strings.TrimSpace(p)
        if t != "" {
            tags = append(tags, t)
        }
    }
    return tags, nil
}

// GenerateEmbedding generates a placeholder embedding using Gemini API
func (g *GeminiService) GenerateEmbedding(text string) ([]float64, error) {
    // Optionally call Gemini for analysis to derive an embedding; placeholder for now
    _, err := g.generateText("Summarize the semantic content for embedding purposes:\n\n"+text, 0.1, 200)
    if err != nil {
        return nil, err
    }
    return []float64{0.1, 0.2, 0.3, 0.4, 0.5}, nil
}

func (g *GeminiService) generateText(prompt string, temperature float64, maxTokens int) (string, error) {
    if len(prompt) > 30000 {
        prompt = prompt[:30000] + "..."
    }

    request := models.GeminiRequest{
        Contents: []models.GeminiContent{{
            Parts: []models.GeminiPart{{Text: prompt}},
        }},
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
    var textParts []string
    for _, part := range response.Candidates[0].Content.Parts {
        textParts = append(textParts, part.Text)
    }
    return strings.Join(textParts, ""), nil
}


