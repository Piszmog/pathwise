package llm

// JobPosting represents the expected JSON structure from Gemini
type JobPosting struct {
	ID                  int64               `json:"id"`
	IsJobPosting        bool                `json:"is_job_posting"`
	CompanyName         string              `json:"company_name"`
	CompanyDescription  string              `json:"company_description"`
	CompanyURL          string              `json:"company_url"`
	JobsURL             string              `json:"jobs_url"`
	ContactEmail        string              `json:"contact_email"`
	Jobs                []Job               `json:"jobs"`
	IsHybrid            bool                `json:"is_hybrid"`
	IsRemote            bool                `json:"is_remote"`
	Location            string              `json:"location"`
	GeneralCompensation GeneralCompensation `json:"general_compensation"`
	GeneralTechStack    []string            `json:"general_tech_stack"`
}

type Job struct {
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	ApplicationURL string       `json:"application_url"`
	RoleType       string       `json:"role_type"`
	Compensation   Compensation `json:"compensation"`
	TechStack      []string     `json:"tech_stack"`
}

type Compensation struct {
	BaseSalary string `json:"base_salary"`
	Equity     string `json:"equity"`
	Other      string `json:"other"`
}

type GeneralCompensation struct {
	BaseSalary string `json:"base_salary"`
	Equity     string `json:"equity"`
}
