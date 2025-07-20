//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestStatus_AllValidTransitions(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "status")
	signin(t, user.Email, "password")

	addJobApplication(t, "Status Test Company", "Software Engineer", "https://statustest.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Test all valid status transitions
	statuses := []string{"watching", "applied", "interviewing", "offered", "accepted"}
	for _, status := range statuses {
		updateJobApplication(t, "", "", "", status)
		require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToHaveValue(status))
		require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText(status)).ToHaveCount(1))
	}
}

func TestStatus_RejectionStatuses(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "status")
	signin(t, user.Email, "password")

	addJobApplication(t, "Rejection Test Company", "Backend Developer", "https://rejectiontest.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Test rejection-type statuses
	rejectionStatuses := []string{"rejected", "declined", "withdrawn", "canceled", "closed"}
	for _, status := range rejectionStatuses {
		updateJobApplication(t, "", "", "", status)
		require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToHaveValue(status))
		require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText(status)).ToHaveCount(1))
	}
}

func TestStatus_HistoryCreation(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "status")
	signin(t, user.Email, "password")

	addJobApplication(t, "History Test Company", "DevOps Engineer", "https://historytest.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Initial status should be "applied"
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("applied")).ToHaveCount(1))

	// Change to interviewing
	updateJobApplication(t, "", "", "", "interviewing")
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("applied")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("interviewing")).ToHaveCount(1))

	// Change to offered
	updateJobApplication(t, "", "", "", "offered")
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("applied")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("interviewing")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("offered")).ToHaveCount(1))

	// Change to accepted
	updateJobApplication(t, "", "", "", "accepted")
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("applied")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("interviewing")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("offered")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("accepted")).ToHaveCount(1))
}

func TestStatus_StatsUpdate(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "status")
	signin(t, user.Email, "password")

	// Add multiple jobs with different statuses
	addJobApplication(t, "Company A", "Job A", "https://companya.com")
	addJobApplication(t, "Company B", "Job B", "https://companyb.com")
	addJobApplication(t, "Company C", "Job C", "https://companyc.com")

	// Initial stats - all applied
	assertStats(t, "3", "3", "0 days", "0%", "0%")

	// Update first job to interviewing
	require.NoError(t, page.GetByText("Company A").Locator("xpath=ancestor::li").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "View job"}).Click())
	updateJobApplication(t, "", "", "", "interviewing")
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Pathwise"}).Click())

	// Stats should show 33% interview rate
	assertStats(t, "3", "3", "0 days", "33%", "0%")

	// Update second job to rejected
	require.NoError(t, page.GetByText("Company B").Locator("xpath=ancestor::li").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "View job"}).Click())
	updateJobApplication(t, "", "", "", "rejected")
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Pathwise"}).Click())

	// Stats should show 33% interview rate, 33% rejection rate
	assertStats(t, "3", "3", "0 days", "33%", "33%")

	// Update third job to offered
	require.NoError(t, page.GetByText("Company C").Locator("xpath=ancestor::li").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "View job"}).Click())
	updateJobApplication(t, "", "", "", "offered")
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Pathwise"}).Click())

	// Stats should show 67% interview rate (interviewing + offered), 33% rejection rate
	assertStats(t, "3", "3", "0 days", "67%", "33%")
}

func TestStatus_FilterByStatus(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "status")
	signin(t, user.Email, "password")

	// Add jobs with different statuses
	addJobApplication(t, "Applied Company", "Job 1", "https://applied.com")
	addJobApplication(t, "Interview Company", "Job 2", "https://interview.com")
	addJobApplication(t, "Rejected Company", "Job 3", "https://rejected.com")

	// Update statuses
	require.NoError(t, page.GetByText("Interview Company").Locator("xpath=ancestor::li").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "View job"}).Click())
	updateJobApplication(t, "", "", "", "interviewing")
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Pathwise"}).Click())

	require.NoError(t, page.GetByText("Rejected Company").Locator("xpath=ancestor::li").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "View job"}).Click())
	updateJobApplication(t, "", "", "", "rejected")
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Pathwise"}).Click())

	// Test filtering by applied status
	filterByStatus(t, "applied")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Applied Company")).ToBeVisible())

	// Test filtering by interviewing status
	filterByStatus(t, "interviewing")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Interview Company")).ToBeVisible())

	// Test filtering by rejected status
	filterByStatus(t, "rejected")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Rejected Company")).ToBeVisible())

	// Clear filter to see all jobs
	clearFilter(t)
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(3))
}

func TestStatus_ArchivedJobStatusDisabled(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "status")
	signin(t, user.Email, "password")

	addJobApplication(t, "Archive Status Test", "Software Engineer", "https://archivestatus.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Verify status select is enabled initially
	require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToBeEnabled())

	// Archive the job
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"}).Click())
	waitForHTMXRequest(t)

	// Navigate to archives
	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)

	// Open archived job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Verify status select is disabled for archived job
	require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToBeDisabled())
}

func TestStatus_UnarchiveJobStatusEnabled(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "status")
	signin(t, user.Email, "password")

	addJobApplication(t, "Unarchive Status Test", "Backend Developer", "https://unarchivestatus.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Archive the job
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"}).Click())
	waitForHTMXRequest(t)

	// Navigate to archives
	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)

	// Open archived job details and unarchive
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Unarchive"}).Click())
	waitForHTMXRequest(t)

	// Navigate back to main page
	_, err = page.Goto(getFullPath(""))
	require.NoError(t, err)

	// Open unarchived job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Verify status select is enabled again
	require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToBeEnabled())

	// Test that status can be updated
	updateJobApplication(t, "", "", "", "interviewing")
	require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToHaveValue("interviewing"))
}

func TestStatus_StatusPersistenceAcrossPageRefresh(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "status")
	signin(t, user.Email, "password")

	addJobApplication(t, "Persistence Test Company", "Full Stack Developer", "https://persistencetest.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Update status to interviewing
	updateJobApplication(t, "", "", "", "interviewing")
	require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToHaveValue("interviewing"))

	// Refresh the page
	_, err := page.Reload()
	require.NoError(t, err)

	// Re-signin after refresh
	signin(t, user.Email, "password")

	// Open job details again
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Verify status persisted
	require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToHaveValue("interviewing"))
	require.NoError(t, expect.Locator(page.Locator("#timeline").GetByText("interviewing")).ToHaveCount(1))
}
