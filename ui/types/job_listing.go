package types

type JobListing struct {
	ID                 string
	Source             JobSource
	SourceID           string
	SourceURL          *string
	Company            string
	CompanyDescription string
	Title              string
	CompanyURL         *string
	ContactEmail       *string
	Description        *string
	RoleType           *string
	Location           *string
	Salary             *string
	Equity             *string
	IsHybrid           bool
	IsRemote           bool
}

type JobListingDetails struct {
	JobListing

	TechStacks []string
}
