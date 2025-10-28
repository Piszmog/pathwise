package types

type JobListing struct {
	ID                 string
	Source             JobSource
	SourceID           string
	SourceURL          string
	Company            string
	CompanyDescription string
	ApplicationURL     string
	Title              string
	CompanyURL         string
	ContactEmail       string
	Description        string
	RoleType           string
	Location           string
	Salary             string
	Equity             string
	IsHybrid           bool
	IsRemote           bool
	PostedAt           string
}

type JobListingDetails struct {
	JobListing

	TechStacks []string
	HasAdded   bool
}

type JobListingFilterOpts struct {
	Title     *string
	IsRemote  *bool
	IsHybrid  *bool
	TechStack *string
}
