//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestHome_NewUser(t *testing.T) {
	beforeEach(t)
	signin(t, "user1@email.com", "password")

	assertStats(t, "0", "0", "0 days", "NaN%", "NaN%")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("Showing 0 to 0 of 0 results ")).ToHaveCount(1))
}

func TestHome_AddApplication(t *testing.T) {
	beforeEach(t)
	signin(t, "user2@email.com", "password")

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("Showing 0 to 0 of 0 results")).ToHaveCount(1))

	addJobApplication(t, "Super Company", "Rock Star", "https://supercompany.com")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 1 of 1 results")).ToHaveCount(1))

	assertStats(t, "1", "1", "0 days", "0%", "0%")
}

func TestHome_UpdatedStats(t *testing.T) {
	beforeEach(t)
	signin(t, "user3@email.com", "password")

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	assertStats(t, "1", "1", "0 days", "0%", "0%")

	updateJobApplication(t, "", "", "", "rejected")
	assertStats(t, "1", "1", "2 days", "0%", "100%")
}

func TestHome_UpdateAllFields(t *testing.T) {
	beforeEach(t)
	signin(t, "user4@email.com", "password")

	addJobApplication(t, "Initial Company", "Initial Title", "https://initial.com")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))

	updateJobApplication(t, "Updated Company", "Updated Title", "https://updated.com", "interviewing")

	require.NoError(t, expect.Locator(page.Locator("#job-form #company")).ToHaveValue("Updated Company"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #title")).ToHaveValue("Updated Title"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #url")).ToHaveValue("https://updated.com"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToHaveValue("interviewing"))
}

func TestHome_AddNote(t *testing.T) {
	beforeEach(t)
	signin(t, "user5@email.com", "password")

	addJobApplication(t, "Note Test Company", "Software Engineer", "https://notetest.com")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))

	addNote(t, "Had a great phone screening today. Moving to technical round next week.")
	require.NoError(t, expect.Locator(page.GetByText("Had a great phone screening today. Moving to technical round next week.")).ToHaveCount(1))

	addNote(t, "Completed technical interview. Waiting for feedback.")
	require.NoError(t, expect.Locator(page.GetByText("Had a great phone screening today. Moving to technical round next week.")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Completed technical interview. Waiting for feedback.")).ToHaveCount(1))
}

func TestHome_StatusUpdateTimeline(t *testing.T) {
	beforeEach(t)
	signin(t, "user6@email.com", "password")

	addJobApplication(t, "Timeline Test Company", "Backend Developer", "https://timeline.com")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("applied")).ToHaveCount(1))

	updateJobApplication(t, "", "", "", "interviewing")
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("applied")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("interviewing")).ToHaveCount(1))

	updateJobApplication(t, "", "", "", "offered")
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("applied")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("interviewing")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("offered")).ToHaveCount(1))
}

func TestHome_BulkArchiveByDate(t *testing.T) {
	beforeEach(t)
	signin(t, "user7@email.com", "password")

	addJobApplication(t, "Old Company 1", "Software Engineer", "https://old1.com")
	addJobApplication(t, "Old Company 2", "Backend Developer", "https://old2.com")
	addJobApplication(t, "Recent Company", "Frontend Developer", "https://recent.com")

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(3))
	require.NoError(t, expect.Locator(page.GetByText("3 results")).ToHaveCount(1))
	assertStats(t, "3", "3", "0 days", "0%", "0%")

	tomorrow := "2025-07-19"
	archiveJobsByDate(t, tomorrow)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("0 results")).ToHaveCount(1))
	assertStats(t, "0", "0", "0 days", "NaN%", "NaN%")

	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(3))
	require.NoError(t, expect.Locator(page.GetByText("Old Company 1")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Old Company 2")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Recent Company")).ToHaveCount(1))
}

func TestHome_FilterFunctionality(t *testing.T) {
	beforeEach(t)
	signin(t, "user8@email.com", "password")

	addJobApplication(t, "Google", "Software Engineer", "https://google.com")
	addJobApplication(t, "Microsoft", "Backend Developer", "https://microsoft.com")
	addJobApplication(t, "Apple", "Frontend Developer", "https://apple.com")
	addJobApplication(t, "Amazon", "DevOps Engineer", "https://amazon.com")

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(4))

	filterByCompany(t, "Google")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Google")).ToHaveCount(1))

	clearFilter(t)
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(4))

	filterByCompany(t, "Micro")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Microsoft")).ToHaveCount(1))

	clearFilter(t)
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(4))

	filterByStatus(t, "applied")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(4))

	clearFilter(t)
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(4))

	filterByCompany(t, "NonExistent")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("0 results")).ToHaveCount(1))

	clearFilter(t)
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(4))
}

func TestHome_ArchiveSingleJob(t *testing.T) {
	beforeEach(t)
	signin(t, "user9@email.com", "password")

	addJobApplication(t, "Archive Test Company", "Software Engineer", "https://archivetest.com")
	addJobApplication(t, "Keep This Company", "Backend Developer", "https://keepthis.com")

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(2))
	require.NoError(t, expect.Locator(page.GetByText("2 results")).ToHaveCount(1))
	assertStats(t, "2", "2", "0 days", "0%", "0%")

	archiveSingleJob(t, "Archive Test Company")

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("1 result")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Keep This Company")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Archive Test Company")).ToHaveCount(0))
	assertStats(t, "1", "1", "0 days", "0%", "0%")

	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Archive Test Company")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Keep This Company")).ToHaveCount(0))
}

func signin(t *testing.T, email, password string) {
	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill(email))
	require.NoError(t, page.Locator("#password").Fill(password))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func addJobApplication(t *testing.T, company, title, url string) {
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
	require.NoError(t, page.Locator("#new-job-form #company").Fill(company))
	require.NoError(t, page.Locator("#new-job-form #title").Fill(title))
	require.NoError(t, page.Locator("#new-job-form #url").Fill(url))
	require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())
}

func assertStats(t *testing.T, totalApps, totalCompanies, hearBack, interviewRate, rejectionRate string) {
	require.NoError(t, expect.Locator(page.Locator("#stats div").Locator("#stats-total-applications")).ToHaveText("Total Applications"+totalApps))
	require.NoError(t, expect.Locator(page.Locator("#stats div").Locator("#stats-total-companies")).ToHaveText("Total Companies"+totalCompanies))
	require.NoError(t, expect.Locator(page.Locator("#stats div").Locator("#stats-average-time-to-hear-back")).ToHaveText("Average time to hear back"+hearBack))
	require.NoError(t, expect.Locator(page.Locator("#stats div").Locator("#stats-interview-percentage")).ToHaveText("Interview Rate"+interviewRate))
	require.NoError(t, expect.Locator(page.Locator("#stats div").Locator("#stats-rejection-percentage")).ToHaveText("Rejection Rate"+rejectionRate))
}

func updateJobApplication(t *testing.T, company, title, url, status string) {
	jobForm := page.Locator("#job-form")
	if count, _ := jobForm.Count(); count == 0 {
		require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	}
	if company != "" {
		require.NoError(t, page.Locator("#job-form #company").Fill(company))
	}
	if title != "" {
		require.NoError(t, page.Locator("#job-form #title").Fill(title))
	}
	if url != "" {
		require.NoError(t, page.Locator("#job-form #url").Fill(url))
	}
	if status != "" {
		_, err := page.Locator("#job-form #status-select").SelectOption(playwright.SelectOptionValues{Values: &[]string{status}})
		require.NoError(t, err)
	}
	require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())
}

func addNote(t *testing.T, note string) {
	noteForm := page.GetByPlaceholder("Add a note...")
	if count, _ := noteForm.Count(); count == 0 {
		require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	}
	require.NoError(t, page.GetByPlaceholder("Add a note...").Fill(note))
	require.NoError(t, page.Locator("#note-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())
}

func archiveJobsByDate(t *testing.T, date string) {
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Archive"}).First().Click())
	require.NoError(t, page.Locator("#date").Fill(date))
	require.NoError(t, page.Locator("#archive-jobs-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"}).Click())
}

func filterByCompany(t *testing.T, company string) {
	require.NoError(t, page.Locator("#filter-form #company").Fill(company))
	require.NoError(t, page.Locator("#filter-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Filter"}).Click())
}

func filterByStatus(t *testing.T, status string) {
	_, err := page.Locator("#filter-form #status-select").SelectOption(playwright.SelectOptionValues{Values: &[]string{status}})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#filter-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Filter"}).Click())
}

func filterByCompanyAndStatus(t *testing.T, company, status string) {
	require.NoError(t, page.Locator("#filter-form #company").Fill(company))
	_, err := page.Locator("#filter-form #status-select").SelectOption(playwright.SelectOptionValues{Values: &[]string{status}})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#filter-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Filter"}).Click())
}

func clearFilter(t *testing.T) {
	require.NoError(t, page.Locator("#filter-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Clear"}).Click())
}

func archiveSingleJob(t *testing.T, companyName string) {
	jobRow := page.GetByText(companyName).Locator("xpath=ancestor::li")
	require.NoError(t, jobRow.GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "View job"}).Click())
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"}).Click())

	page.WaitForTimeout(1000)
}
