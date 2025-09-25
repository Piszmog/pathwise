package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

var _ Client = (*GeminiClient)(nil)

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.0-flash-exp")

	model.SetTemperature(0.1) // Low temperature for consistent parsing
	model.ResponseMIMEType = "application/json"

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

func (c *GeminiClient) ParseJobPostings(ctx context.Context, inputs map[int64]string) ([]JobPosting, error) {
	if len(inputs) == 0 {
		return nil, errors.New("no inputs provided")
	}
	if len(inputs) > 10 {
		return nil, errors.New("maximum 10 job postings per batch")
	}

	var jobPostings []JobPosting
	prompt := fmt.Sprintf(batchPromptTemplate, formatInputsForPrompt(inputs))

	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return jobPostings, getRateLimitError(err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return jobPostings, ErrNoResponse
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	if err := json.Unmarshal([]byte(responseText), &jobPostings); err != nil {
		return jobPostings, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return jobPostings, nil
}

// formatInputsForPrompt formats multiple inputs for the batch prompt
func formatInputsForPrompt(inputs map[int64]string) string {
	// Sort IDs to ensure consistent ordering
	var ids []int64
	for id := range inputs {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	var formatted strings.Builder
	for i, id := range ids {
		input := inputs[id]
		formatted.WriteString(fmt.Sprintf("## Text %d (ID: %d):\n", i+1, id))
		formatted.WriteString("```\n")
		formatted.WriteString(input)
		formatted.WriteString("\n```\n")
		if i < len(ids)-1 {
			formatted.WriteString("\n")
		}
	}
	return formatted.String()
}

const batchPromptTemplate = `# Batch Posting Parser Prompt

You are a specialized parser for job posts. Your task is to extract structured job information from multiple texts that may contain job postings or regular comments.

## Input Format
You will receive 1-10 HTML-encoded texts from job posts. Each text has a unique ID that you MUST include in your response. Each text may contain:
- Job postings with company information, roles, locations, and compensation
- Regular comments that are not job postings
- Mixed content with multiple job roles from the same company

## Text Preprocessing Instructions
Before extracting information, normalize the following for each text:

### HTML Entity Decoding
- Decode HTML entities like &#x2F; (/) and &#x27; (') back to regular characters
- Convert URLs like 'https:&#x2F;&#x2F;www.example.com&#x2F;' to 'https://www.example.com/'

### Email Format Normalization
- Convert obfuscated emails to standard format:
  - 'careers<at>company<dot>com' → 'careers@company.com'
  - 'hiring[at]company[dot]com' → 'hiring@company.com'
  - 'contact AT company DOT com' → 'contact@company.com'
  - 'email_hiring_2025 AT company.ai' → 'email_hiring_2025@company.ai'

### Text Case Normalization
- Convert ALL CAPS company names to proper title case (e.g., "ACME CORP" → "Acme Corp")
- Convert ALL CAPS job titles to proper title case (e.g., "SOFTWARE ENGINEER" → "Software Engineer")
- Preserve intentional capitalization in acronyms and technical terms

## Critical Output Requirements
- Return ONLY valid JSON array - no markdown code blocks, no explanations, no additional text
- Process each text independently and return results in the same order as inputs
- MUST include the correct ID for each text in the "id" field
- Do NOT include citations, reference numbers, or bracketed numbers like [1], [2], etc.
- Use ONLY information directly from the provided text
- Do NOT add external knowledge about companies beyond what's in the text
- All field values must be clean strings without citation markers
- Apply all normalization rules above before populating JSON fields

## JSON Structure
Return an array of job posting objects, one for each input text, maintaining the same order:

[
  {
    "id": "int64 - the ID provided for this specific text (REQUIRED)",
    "is_job_posting": "boolean - true if text contains job posting information, false otherwise",
    "company_name": "string - the company name (required, normalized to proper case)",
    "company_description": "string - brief description of what the company does (optional)",
    "company_url": "string - url of the company main website with HTML entities decoded (optional)",
    "contact_email": "string - the contact email for job applications in standard format (optional)",
    "jobs": [
      {
        "title": "string - job title normalized to proper case",
        "description": "string - specific responsibilities/requirements for this role (optional if no specific description for the specific job)",
        "role_type": "string - 'full-time', 'part-time', 'full-time contractor', 'contract', 'internship', or 'unknown'",
        "compensation": {
          "base_salary": "salary range specific to this job (optional)",
          "equity": "equity details specific to this job (optional)",
          "other": "other compensation specific to this job (optional)"
        },
        "tech_stack": [
          "technologies specific to this job (optional)"
        ]
      }
    ],
    "is_hybrid": "boolean - true only if hybrid work is explicitly mentioned AND remote is not",
    "is_remote": "boolean - true if remote/distributed work is mentioned AND hybrid is not",
    "location": "string - office location, 'Remote', or 'unknown'",
    "general_compensation": {
      "base_salary": "salary that applies to all jobs (optional)",
      "equity": "equity that applies to all jobs (optional)"
    },
    "general_tech_stack": [
      "technologies that apply to all jobs or empty array (optional)"
    ]
  }
]

## Texts to Parse:
%s`

func getRateLimitError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case 429:
			return fmt.Errorf("%w: %s", ErrRateLimit, apiErr.Message)
		case 503:
			return fmt.Errorf("%w: %s", ErrServiceUnavailable, apiErr.Message)
		case 403:
			if strings.Contains(strings.ToLower(apiErr.Message), "quota") {
				return fmt.Errorf("%w: %s", ErrQuotaExhausted, apiErr.Message)
			}
		}
		return err
	}

	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "too many requests") {
		return fmt.Errorf("%w: %s", ErrRateLimit, err.Error())
	}
	if strings.Contains(errMsg, "quota exceeded") || strings.Contains(errMsg, "quota") {
		return fmt.Errorf("%w: %s", ErrQuotaExhausted, err.Error())
	}
	if strings.Contains(errMsg, "resource_exhausted") {
		return fmt.Errorf("%w: %s", ErrServiceUnavailable, err.Error())
	}

	return err
}

var (
	ErrNoResponse         = errors.New("no response")
	ErrRateLimit          = errors.New("rate limit exceeded")
	ErrQuotaExhausted     = errors.New("quota exhausted")
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
)

func (c *GeminiClient) Close() error {
	return c.client.Close()
}
