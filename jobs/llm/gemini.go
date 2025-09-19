package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func (c *GeminiClient) ParseJobPosting(ctx context.Context, input string) (JobPosting, error) {
	var jobPosting JobPosting
	prompt := fmt.Sprintf(promptTemplate, input)

	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return jobPosting, getRateLimitError(err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return jobPosting, ErrNoResponse
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	if err := json.Unmarshal([]byte(responseText), &jobPosting); err != nil {
		return jobPosting, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return jobPosting, nil
}

const promptTemplate = `# Posting Parser Prompt

You are a specialized parser for job posts. Your task is to extract structured job information from text that may contain job postings or regular comments.

## Input Format
You will receive HTML-encoded text from job posts that may contain:
- Job postings with company information, roles, locations, and compensation
- Regular comments that are not job postings
- Mixed content with multiple job roles from the same company

## Critical Output Requirements
- Return ONLY valid JSON - no markdown code blocks, no explanations, no additional text
- Do NOT include citations, reference numbers, or bracketed numbers like [1], [2], etc.
- Use ONLY information directly from the provided text
- Do NOT add external knowledge about companies beyond what's in the text
- All field values must be clean strings without citation markers

## JSON Structure
{
  "is_job_posting": "boolean - true if text contains job posting information, false otherwise",
  "company_name": "string - the company name (required)",
  "company_description": "string - brief description of what the company does (optional)",
  "company_url": "string - url of the company main website (optional)",
  "contact_email": "string - the contact email for job applications (optional)",
  "jobs": [
    {
      "title": "string - job title",
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

## Text to Parse:
` + "```\n%s\n```"

func getRateLimitError(err error) error {
	if err == nil {
		return nil
	}

	if apiErr, ok := err.(*googleapi.Error); ok {
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
