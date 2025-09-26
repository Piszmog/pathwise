package types

import (
	"time"
)

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
	CreatedAt          time.Time
}

func GetMockJobListings() []JobListing {
	companyURL1 := "https://example.com"
	companyURL2 := "https://startup.io"
	companyURL3 := "https://blockchain-labs.com"
	sourceURL1 := "https://news.ycombinator.com/item?id=12345"
	sourceURL2 := "https://linkedin.com/jobs/12346"
	sourceURL3 := "https://indeed.com/viewjob?jk=12347"
	location1 := "San Francisco, CA"
	location2 := "Remote"
	location3 := "New York, NY"
	location4 := "Austin, TX"
	salary1 := "$120k - $180k"
	salary2 := "$100k - $150k"
	salary3 := "$90k - $140k"
	salary4 := "$80k - $120k"
	equity1 := "0.1% - 0.5%"
	equity2 := "0.05% - 0.2%"
	roleType1 := "Full-time"
	roleType2 := "Contract"
	contactEmail1 := "jobs@techcorp.com"
	contactEmail2 := "careers@aistartup.io"
	contactEmail3 := "hiring@blockchain-labs.com"
	description1 := "We're looking for a senior software engineer to join our growing team."
	description2 := "Join our AI startup and help build the future of technology."
	description3 := "Work on cutting-edge blockchain technology with our distributed team."
	description4 := "Build developer tools that make engineers more productive."

	return []JobListing{
		{
			ID:                 "job-listing-1",
			Source:             JobSourceHackerNews,
			SourceID:           "12345",
			SourceURL:          &sourceURL1,
			Company:            "TechCorp",
			CompanyDescription: "Leading technology company building innovative solutions",
			Title:              "Senior Software Engineer",
			CompanyURL:         &companyURL1,
			Location:           &location1,
			Salary:             &salary1,
			Equity:             &equity1,
			RoleType:           &roleType1,
			ContactEmail:       &contactEmail1,
			Description:        &description1,
			IsHybrid:           true,
			IsRemote:           false,
			CreatedAt:          time.Now().AddDate(0, 0, -2),
		},
		{
			ID:                 "job-listing-2",
			Source:             JobSourceLinkedIn,
			SourceID:           "linkedin-12346",
			SourceURL:          &sourceURL2,
			Company:            "AI Startup",
			CompanyDescription: "Cutting-edge AI company revolutionizing machine learning",
			Title:              "Frontend Developer",
			CompanyURL:         &companyURL2,
			Location:           &location2,
			Salary:             &salary2,
			Equity:             &equity2,
			RoleType:           &roleType1,
			ContactEmail:       &contactEmail2,
			Description:        &description2,
			IsHybrid:           false,
			IsRemote:           true,
			CreatedAt:          time.Now().AddDate(0, 0, -1),
		},
		{
			ID:                 "job-listing-3",
			Source:             JobSourceIndeed,
			SourceID:           "indeed-12347",
			SourceURL:          &sourceURL3,
			Company:            "Blockchain Labs",
			CompanyDescription: "Decentralized technology platform for the future",
			Title:              "Backend Engineer",
			CompanyURL:         &companyURL3,
			Location:           &location3,
			Salary:             &salary3,
			RoleType:           &roleType2,
			ContactEmail:       &contactEmail3,
			Description:        &description3,
			IsHybrid:           false,
			IsRemote:           false,
			CreatedAt:          time.Now().AddDate(0, 0, -3),
		},
		{
			ID:                 "job-listing-4",
			Source:             JobSourceAngelList,
			SourceID:           "angellist-12348",
			Company:            "DevTools Inc",
			CompanyDescription: "Building developer productivity tools",
			Title:              "Full Stack Developer",
			Location:           &location2,
			Salary:             &salary4,
			RoleType:           &roleType1,
			Description:        &description4,
			IsHybrid:           false,
			IsRemote:           true,
			CreatedAt:          time.Now().AddDate(0, 0, -5),
		},
		{
			ID:                 "job-listing-5",
			Source:             JobSourceHackerNews,
			SourceID:           "hn-12349",
			Company:            "DataCorp",
			CompanyDescription: "Big data analytics and machine learning platform",
			Title:              "Data Engineer",
			Location:           &location4,
			Salary:             &salary1,
			RoleType:           &roleType1,
			IsHybrid:           true,
			IsRemote:           false,
			CreatedAt:          time.Now().AddDate(0, 0, -7),
		},
		{
			ID:                 "job-listing-6",
			Source:             JobSourceLinkedIn,
			SourceID:           "linkedin-12350",
			Company:            "FlexiWork Inc",
			CompanyDescription: "Modern workplace solutions company",
			Title:              "Product Manager",
			Location:           &location2,
			Salary:             &salary2,
			RoleType:           &roleType1,
			Description:        &description4,
			IsHybrid:           true,
			IsRemote:           true,
			CreatedAt:          time.Now().AddDate(0, 0, -4),
		},
	}
}

type JobListingDetails struct {
	JobListing

	TechStacks []string
}

func GetMockJobListingDetails(id string) *JobListingDetails {
	jobs := GetMockJobListings()
	for _, job := range jobs {
		if job.ID == id {
			return &JobListingDetails{
				JobListing: job,
				TechStacks: getMockTechStacks(id),
			}
		}
	}
	return nil
}

func getMockTechStacks(jobID string) []string {
	techStacks := map[string][]string{
		"job-listing-1": {"Go", "React", "TypeScript", "PostgreSQL", "AWS"},
		"job-listing-2": {"JavaScript", "Vue.js", "Node.js", "MongoDB", "Docker"},
		"job-listing-3": {"Rust", "Blockchain", "Solidity", "Ethereum", "Docker"},
		"job-listing-4": {"Python", "Django", "React", "PostgreSQL", "Redis"},
		"job-listing-5": {"Python", "Apache Spark", "Kafka", "Snowflake", "AWS"},
		"job-listing-6": {"React", "Node.js", "GraphQL", "MongoDB", "Kubernetes"},
	}
	if stacks, exists := techStacks[jobID]; exists {
		return stacks
	}
	return []string{}
}
