package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"hub-hrms/backend/internal/service"
)

// RegisterAIRoutes registers AI-related routes
func RegisterAIRoutes(r chi.Router, services *service.Services) {
	r.Route("/ai", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Post("/generate-job-description", generateJobDescriptionHandler(services))
	})
}

// AIGenerateJobRequest represents the request for AI job generation
type AIGenerateJobRequest struct {
	Title            string   `json:"title"`
	Department       string   `json:"department"`
	Location         string   `json:"location"`
	EmploymentType   string   `json:"employment_type"`
	SalaryMin        float64  `json:"salary_min"`
	SalaryMax        float64  `json:"salary_max"`
	Description      string   `json:"description"`
	Requirements     []string `json:"requirements"`
	Responsibilities []string `json:"responsibilities"`
	Benefits         []string `json:"benefits"`
	VersionType      string   `json:"version_type"` // "jobboard" or "internal"
}

// AIGenerateJobResponse represents the AI generation response
type AIGenerateJobResponse struct {
	GeneratedText string `json:"generated_text"`
	VersionType   string `json:"version_type"`
}

// AnthropicMessage represents a message in the Anthropic API format
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicRequest represents the request to Anthropic API
type AnthropicRequest struct {
	Model      string             `json:"model"`
	MaxTokens  int                `json:"max_tokens"`
	Messages   []AnthropicMessage `json:"messages"`
}

// AnthropicContentBlock represents a content block in the response
type AnthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicResponse represents the response from Anthropic API
type AnthropicResponse struct {
	Content []AnthropicContentBlock `json:"content"`
	Model   string                  `json:"model"`
	Role    string                  `json:"role"`
}

func generateJobDescriptionHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req AIGenerateJobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate required fields
		if req.Title == "" || req.Description == "" {
			respondError(w, http.StatusBadRequest, "title and description are required")
			return
		}

		// Build job context
		jobContext := buildJobContext(req)

		// Build prompt based on version type
		var prompt string
		if req.VersionType == "jobboard" {
			prompt = buildJobBoardPrompt(jobContext)
		} else if req.VersionType == "internal" {
			prompt = buildInternalPrompt(jobContext)
		} else {
			respondError(w, http.StatusBadRequest, "version_type must be 'jobboard' or 'internal'")
			return
		}

		// Call Anthropic API
		generatedText, err := callAnthropicAPI(prompt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate description: %v", err))
			return
		}

		response := AIGenerateJobResponse{
			GeneratedText: generatedText,
			VersionType:   req.VersionType,
		}

		respondJSON(w, http.StatusOK, response)
	}
}

func buildJobContext(req AIGenerateJobRequest) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("Job Title: %s\n", req.Title))
	sb.WriteString(fmt.Sprintf("Department: %s\n", req.Department))
	sb.WriteString(fmt.Sprintf("Location: %s\n", req.Location))
	sb.WriteString(fmt.Sprintf("Employment Type: %s\n", req.EmploymentType))
	
	if req.SalaryMin > 0 && req.SalaryMax > 0 {
		sb.WriteString(fmt.Sprintf("Salary Range: $%.0f - $%.0f\n", req.SalaryMin, req.SalaryMax))
	}
	
	sb.WriteString(fmt.Sprintf("\nBase Description:\n%s\n", req.Description))
	
	if len(req.Requirements) > 0 {
		sb.WriteString("\nRequirements:\n")
		for _, r := range req.Requirements {
			if strings.TrimSpace(r) != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", r))
			}
		}
	}
	
	if len(req.Responsibilities) > 0 {
		sb.WriteString("\nResponsibilities:\n")
		for _, r := range req.Responsibilities {
			if strings.TrimSpace(r) != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", r))
			}
		}
	}
	
	if len(req.Benefits) > 0 {
		sb.WriteString("\nBenefits:\n")
		for _, b := range req.Benefits {
			if strings.TrimSpace(b) != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", b))
			}
		}
	}
	
	return sb.String()
}

func buildJobBoardPrompt(jobContext string) string {
	return fmt.Sprintf(`Based on the following job information, create an engaging, concise job posting optimized for external job boards (LinkedIn, Indeed, etc.).

%s

Requirements for Job Board Version:
- Keep it concise (300-400 words)
- Use engaging, action-oriented language
- Focus on selling the opportunity
- Highlight key benefits and perks
- Include company culture elements
- Use a professional but friendly tone
- Make it scannable with short paragraphs
- End with a clear call-to-action

Format the output as a ready-to-post job description. Do not include any preamble or explanatory text, just the job description itself.`, jobContext)
}

func buildInternalPrompt(jobContext string) string {
	return fmt.Sprintf(`Based on the following job information, create a detailed, comprehensive job description for internal HR use.

%s

Requirements for Internal HR Version:
- Be comprehensive and detailed (500-700 words)
- Include full job specifications
- Detail all requirements and qualifications
- List all responsibilities with specifics
- Include performance expectations
- Add compliance and legal considerations
- Use formal, professional language
- Include evaluation criteria
- Provide context for hiring managers

Format the output as a formal internal job description document. Do not include any preamble or explanatory text, just the job description itself.`, jobContext)
}

func callAnthropicAPI(prompt string) (string, error) {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY environment variable not set. Get your API key from https://console.anthropic.com/")
	}

	// Create request payload
	reqBody := AnthropicRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1000,
		Messages: []AnthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, string(body))
	}

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text from content blocks
	var generatedText strings.Builder
	for _, content := range anthropicResp.Content {
		if content.Type == "text" {
			generatedText.WriteString(content.Text)
		}
	}

	return generatedText.String(), nil
}