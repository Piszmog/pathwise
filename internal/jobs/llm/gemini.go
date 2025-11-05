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

	model := client.GenerativeModel("gemini-2.5-flash")

	model.SetTemperature(0.1) // Low temperature for consistent parsing
	model.ResponseMIMEType = "application/json"

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

func (c *GeminiClient) ParseJobPostings(ctx context.Context, inputs map[int64]string) ([]JobPosting, error) {
	if len(inputs) == 0 {
		return nil, ErrNoInputs
	}
	if len(inputs) > 30 {
		return nil, ErrMaxBatch
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

var ErrNoInputs = errors.New("no inputs provided")
var ErrMaxBatch = errors.New("maximum 30 job postings per batch")

func formatInputsForPrompt(inputs map[int64]string) string {
	var ids = make([]int64, 0, len(inputs))
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

const batchPromptTemplate = `Parse job postings from HTML text and extract structured information.

## Input Format
You will receive 1-30 HTML-encoded texts with unique IDs. Each may contain job postings, regular comments, or mixed content with multiple roles from the same company.

## Text Preprocessing

### HTML Entity Decoding
Decode HTML entities in URLs and text:
- &#x2F; → /
- &#x27; → '
- &#x3A; → :
- &#x3D; → =
- &#x3F; → ?
- &#x26; → &

Example: 'https:&#x2F;&#x2F;example.com&#x2F;apply' → 'https://example.com/apply'

### Email Normalization
Convert obfuscated emails to standard format:
- 'contact<at>company<dot>com' → 'contact@company.com'
- 'hiring AT company DOT com' → 'hiring@company.com'

### URL Extraction (CRITICAL)
- Extract FULL URLs from <a href="..."> tags, NOT the display text
- Example: <a href="https://example.com/apply">https://example.com/ap...</a> → use 'https://example.com/apply'
- For plain text URLs truncated with "..." (like https://example.com/apply?id=abc...), extract the URL AS-IS including the "..."
- Decode all HTML entities in URLs
- Prioritize URLs with keywords: 'apply', 'application', 'jobs', 'careers', 'positions', 'hiring'

**jobs_url vs application_url:**
- Use jobs_url if one URL applies to all jobs in the posting
- Use application_url within each job object if URLs are role-specific
- When uncertain, prefer jobs_url if only one URL is provided

### Text Normalization
- Convert ALL CAPS to proper title case: "ACME CORP" → "Acme Corp"
- Preserve acronyms and technical terms (API, SQL, AWS, etc.)

### Company Description Extraction
Extract company information from the posting:
- Look for sentences describing what the company does, builds, or works on
- Include industry, products, customers, or mission statements
- Combine related sentences into a concise 1-3 sentence summary
- Exclude information specific to individual job roles
- Examples of what to capture:
  - "We build tools for the mortgage industry"
  - "Work with Fortune 500 companies"
  - "At the forefront of applying AI in healthcare"
- Keep descriptions factual and relevant to job seekers

### Job Description Extraction
For each job role, extract:
- What the person will do/work on (responsibilities)
- What the role involves (day-to-day activities)
- Who they'll work with (team, stakeholders)
- What impact they'll have
- Key requirements or qualifications mentioned
- Combine into a coherent 1-3 sentence summary per job
- Exclude generic information that applies to all jobs
- Examples:
  - "Working directly with CTO on document processing systems. New-grad and junior engineers encouraged."
  - "Manage AWS infrastructure with Terraform, handle software engineering tasks, and oversee IT operations."

### Technology Stack Extraction (CRITICAL)
Extract technologies from job titles AND descriptions:
- Job titles: "Senior Go Engineer" → tech_stack: ["Go"]
- Descriptions: look for "tech stack:", "We use:", "experience with:", "technologies:", "working with"
- Include: programming languages, frameworks, databases, cloud platforms, tools, protocols
- Normalize names: "golang" → "Go", "reactjs" → "React", "postgres" → "PostgreSQL", "aws" → "AWS"
- Include AI/ML terms: "LLM" → "LLM", "machine learning" → "Machine Learning"

**general_tech_stack vs tech_stack:**
- Use general_tech_stack for technologies mentioned at company level or applying to all jobs
- Use tech_stack for job-specific technologies
- If unclear, put technologies in the specific job's tech_stack

### Work Location Logic
- is_remote: true if the posting mentions remote work for ANY of the jobs
- is_hybrid: true if the posting mentions hybrid work for ANY of the jobs
- Both can be true if the posting has mixed work arrangements (some hybrid, some remote)
- location field: capture the primary location info from the posting header or description
  - If multiple locations mentioned, use the most general description (e.g., "Hybrid SF & Remote", "US Remote", "SF Bay Area")
  - If specific cities for different roles, prefer the company headquarters or first mentioned location

### Compensation Handling
- Put compensation in general_compensation if it applies to all jobs
- Put compensation in individual job's compensation field if it's role-specific
- If a salary range is given in the header (e.g., "$100k-$220k"), assume it applies to all jobs

## Output Requirements
- Return ONLY valid JSON array (no markdown, no explanations, no extra text)
- Process texts in order with correct IDs
- NO citations, reference numbers, or brackets like [1], [2]
- Use ONLY information from the provided text
- Do NOT add external knowledge about companies
- Keep descriptions concise but informative (1-3 sentences each)

## JSON Structure
Return an array of objects in the same order as inputs:

[
  {
    "id": 123,
    "is_job_posting": true,
    "company_name": "string (REQUIRED, proper case)",
    "company_description": "string (1-3 sentences about what the company does, optional)",
    "company_url": "string (optional, decoded)",
    "contact_email": "string (optional, normalized)",
    "jobs_url": "string (optional, decoded URL for all jobs)",
    "jobs": [
      {
        "title": "string (REQUIRED, proper case)",
        "description": "string (1-3 sentences about role responsibilities and requirements, optional)",
        "role_type": "full-time|part-time|full-time contractor|contract|internship|unknown (REQUIRED)",
        "application_url": "string (optional, decoded URL for this specific job)",
        "compensation": {
          "base_salary": "string (optional, for this specific job only)",
          "equity": "string (optional, for this specific job only)",
          "other": "string (optional, for this specific job only)"
        },
        "tech_stack": ["string array (optional, from title AND description)"]
      }
    ],
    "is_hybrid": false,
    "is_remote": true,
    "location": "string (REQUIRED, 'Remote' or general location description or 'unknown')",
    "general_compensation": {
      "base_salary": "string (optional, applies to all jobs)",
      "equity": "string (optional, applies to all jobs)"
    },
    "general_tech_stack": ["string array (optional, applies to all jobs)"]
  }
]

Notes:
- Include all fields even if empty/null
- Preserve original salary format (e.g., "$100k-$220k", "€50-70k", "120000-150000")
- Descriptions should be concise summaries, not verbatim text dumps

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
