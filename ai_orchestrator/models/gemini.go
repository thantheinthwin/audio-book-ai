package models

// Gemini API request/response types

type GeminiPart struct {
    Text string `json:"text,omitempty"`
}

type GeminiContent struct {
    Parts []GeminiPart `json:"parts"`
}

type GeminiGenerationConfig struct {
    Temperature     float64 `json:"temperature,omitempty"`
    MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type GeminiRequest struct {
    Contents         []GeminiContent        `json:"contents"`
    GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiCandidate struct {
    Content GeminiContent `json:"content"`
}

type GeminiResponse struct {
    Candidates []GeminiCandidate `json:"candidates"`
}


