//go:build e2e

package e2e_test

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestPagination_BasicNavigation(t *testing.T) {
	beforeEach(t)
	signin(t, "pagination_user1@email.com", "password")

	// Add 15 job applications to trigger pagination (default page size is 10)
	for i := 1; i <= 15; i++ {
		addJobApplication(t, fmt.Sprintf("Company %d", i), fmt.Sprintf("Engineer %d", i), fmt.Sprintf("https://company%d.com", i))
	}

	// Verify first page shows 10 items
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 15 results")).ToHaveCount(1))

	// Verify Previous button is disabled on first page
	previousButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Previous"})
	require.NoError(t, expect.Locator(previousButton).ToBeDisabled())

	// Verify Next button is enabled
	nextButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Next"})
	require.NoError(t, expect.Locator(nextButton).ToBeEnabled())

	// Click Next to go to page 2
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	// Verify second page shows remaining 5 items
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(5))
	require.NoError(t, expect.Locator(page.GetByText("Showing 11 to 15 of 15 results")).ToHaveCount(1))

	// Verify Previous button is now enabled
	require.NoError(t, expect.Locator(previousButton).ToBeEnabled())

	// Verify Next button is disabled on last page
	require.NoError(t, expect.Locator(nextButton).ToBeDisabled())

	// Click Previous to go back to page 1
	require.NoError(t, previousButton.Click())
	waitForHTMXRequest(t)

	// Verify we're back on first page
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 15 results")).ToHaveCount(1))
}

func TestPagination_LargeDataset(t *testing.T) {
	beforeEach(t)
	signin(t, "pagination_user2@email.com", "password")

	// Add 25 job applications to test multiple pages
	for i := 1; i <= 25; i++ {
		addJobApplication(t, fmt.Sprintf("BigCorp %d", i), fmt.Sprintf("Role %d", i), fmt.Sprintf("https://bigcorp%d.com", i))
	}

	// Verify first page
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 25 results")).ToHaveCount(1))

	nextButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Next"})
	previousButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Previous"})

	// Navigate to page 2
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 11 to 20 of 25 results")).ToHaveCount(1))

	// Navigate to page 3 (last page)
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(5))
	require.NoError(t, expect.Locator(page.GetByText("Showing 21 to 25 of 25 results")).ToHaveCount(1))

	// Verify Next button is disabled on last page
	require.NoError(t, expect.Locator(nextButton).ToBeDisabled())

	// Navigate back through pages
	require.NoError(t, previousButton.Click())
	waitForHTMXRequest(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 11 to 20 of 25 results")).ToHaveCount(1))

	require.NoError(t, previousButton.Click())
	waitForHTMXRequest(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 25 results")).ToHaveCount(1))

	// Verify Previous button is disabled on first page
	require.NoError(t, expect.Locator(previousButton).ToBeDisabled())
}

func TestPagination_WithFilters(t *testing.T) {
	beforeEach(t)
	signin(t, "pagination_user3@email.com", "password")

	// Add 15 jobs with different companies
	for i := 1; i <= 10; i++ {
		addJobApplication(t, "Google", fmt.Sprintf("Engineer %d", i), fmt.Sprintf("https://google.com/job%d", i))
	}
	for i := 1; i <= 8; i++ {
		addJobApplication(t, "Microsoft", fmt.Sprintf("Developer %d", i), fmt.Sprintf("https://microsoft.com/job%d", i))
	}

	// Apply filter for Google jobs
	filterByCompany(t, "Google")

	// Should show 10 Google jobs on first page
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 10 results")).ToHaveCount(1))

	// Verify all displayed jobs are from Google
	require.NoError(t, expect.Locator(page.GetByText("Google")).ToHaveCount(10))

	// Clear filter and apply Microsoft filter
	clearFilter(t)
	filterByCompany(t, "Microsoft")

	// Should show 8 Microsoft jobs on first page
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(8))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 8 of 8 results")).ToHaveCount(1))

	// Verify all displayed jobs are from Microsoft
	require.NoError(t, expect.Locator(page.GetByText("Microsoft")).ToHaveCount(8))

	// Clear filter to see all jobs again
	clearFilter(t)

	// Should show all 18 jobs with pagination
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 18 results")).ToHaveCount(1))

	// Navigate to second page
	nextButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Next"})
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(8))
	require.NoError(t, expect.Locator(page.GetByText("Showing 11 to 18 of 18 results")).ToHaveCount(1))
}

func TestPagination_FilterPreservesPage(t *testing.T) {
	beforeEach(t)
	signin(t, "pagination_user4@email.com", "password")

	// Add 25 jobs with mixed companies
	for i := 1; i <= 15; i++ {
		addJobApplication(t, "TechCorp", fmt.Sprintf("Engineer %d", i), fmt.Sprintf("https://techcorp.com/job%d", i))
	}
	for i := 1; i <= 10; i++ {
		addJobApplication(t, "StartupInc", fmt.Sprintf("Developer %d", i), fmt.Sprintf("https://startup.com/job%d", i))
	}

	// Navigate to page 2
	nextButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Next"})
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	// Verify we're on page 2
	require.NoError(t, expect.Locator(page.GetByText("Showing 11 to 20 of 25 results")).ToHaveCount(1))

	// Apply filter - should reset to page 1
	filterByCompany(t, "TechCorp")

	// Should show TechCorp jobs starting from page 1
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 15 results")).ToHaveCount(1))

	// Navigate to page 2 of filtered results
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(5))
	require.NoError(t, expect.Locator(page.GetByText("Showing 11 to 15 of 15 results")).ToHaveCount(1))

	// Clear filter - should reset to page 1 again
	clearFilter(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 25 results")).ToHaveCount(1))
}

func TestPagination_EdgeCases(t *testing.T) {
	beforeEach(t)
	signin(t, "pagination_user5@email.com", "password")

	// Test with exactly 10 items (no pagination needed)
	for i := 1; i <= 10; i++ {
		addJobApplication(t, fmt.Sprintf("Company %d", i), fmt.Sprintf("Role %d", i), fmt.Sprintf("https://company%d.com", i))
	}

	// Should show all 10 items on one page
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 10 results")).ToHaveCount(1))

	// Both pagination buttons should be disabled
	previousButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Previous"})
	nextButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Next"})

	require.NoError(t, expect.Locator(previousButton).ToBeDisabled())
	require.NoError(t, expect.Locator(nextButton).ToBeDisabled())

	// Add one more item to trigger pagination
	addJobApplication(t, "Company 11", "Role 11", "https://company11.com")

	// Should now show pagination
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 11 results")).ToHaveCount(1))

	// Previous should be disabled, Next should be enabled
	require.NoError(t, expect.Locator(previousButton).ToBeDisabled())
	require.NoError(t, expect.Locator(nextButton).ToBeEnabled())

	// Navigate to last page
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	// Should show 1 item on page 2
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Showing 11 to 11 of 11 results")).ToHaveCount(1))

	// Previous should be enabled, Next should be disabled
	require.NoError(t, expect.Locator(previousButton).ToBeEnabled())
	require.NoError(t, expect.Locator(nextButton).ToBeDisabled())
}

func TestPagination_EmptyResults(t *testing.T) {
	beforeEach(t)
	signin(t, "pagination_user6@email.com", "password")

	// Start with no jobs
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("Showing 0 to 0 of 0 results")).ToHaveCount(1))

	// Pagination should not be visible or buttons should be disabled
	previousButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Previous"})
	nextButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Next"})

	require.NoError(t, expect.Locator(previousButton).ToBeDisabled())
	require.NoError(t, expect.Locator(nextButton).ToBeDisabled())

	// Add some jobs then filter to get empty results
	for i := 1; i <= 5; i++ {
		addJobApplication(t, "RealCompany", fmt.Sprintf("Role %d", i), fmt.Sprintf("https://real%d.com", i))
	}

	// Filter for non-existent company
	filterByCompany(t, "NonExistentCompany")

	// Should show no results
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("Showing 0 to 0 of 0 results")).ToHaveCount(1))

	// Pagination buttons should be disabled
	require.NoError(t, expect.Locator(previousButton).ToBeDisabled())
	require.NoError(t, expect.Locator(nextButton).ToBeDisabled())
}

func TestPagination_StatusFilter(t *testing.T) {
	beforeEach(t)
	signin(t, "pagination_user7@email.com", "password")

	// Add 12 jobs - this will create pagination
	for i := 1; i <= 12; i++ {
		addJobApplication(t, fmt.Sprintf("Company %d", i), fmt.Sprintf("Role %d", i), fmt.Sprintf("https://company%d.com", i))
	}

	// Verify we have pagination initially
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 12 results")).ToHaveCount(1))

	// Update one job to "interviewing" status
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	waitForHTMXRequest(t)

	_, err := page.Locator("#job-form #status-select").SelectOption(playwright.SelectOptionValues{Values: &[]string{"interviewing"}})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())
	waitForHTMXRequest(t)

	// Go back to home page
	_, err = page.Goto(getFullPath(""))
	require.NoError(t, err)

	// Filter by "interviewing" status
	filterByStatus(t, "interviewing")

	// Should show 1 interviewing job
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 1 of 1 results")).ToHaveCount(1))

	// Filter by "applied" status
	filterByStatus(t, "applied")

	// Should show 11 applied jobs with pagination
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 11 results")).ToHaveCount(1))

	// Next button should be enabled (more than 10 items)
	nextButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Next"})
	require.NoError(t, expect.Locator(nextButton).ToBeEnabled())

	// Navigate to page 2
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	// Should show 1 item on page 2
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Showing 11 to 11 of 11 results")).ToHaveCount(1))
}

func TestPagination_AddJobUpdatesCount(t *testing.T) {
	beforeEach(t)
	signin(t, "pagination_user8@email.com", "password")

	// Add 10 jobs (exactly one page)
	for i := 1; i <= 10; i++ {
		addJobApplication(t, fmt.Sprintf("Company %d", i), fmt.Sprintf("Role %d", i), fmt.Sprintf("https://company%d.com", i))
	}

	// Verify pagination shows 10 items, no next button
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 10 results")).ToHaveCount(1))

	nextButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Next"})
	require.NoError(t, expect.Locator(nextButton).ToBeDisabled())

	// Add one more job
	addJobApplication(t, "Company 11", "Role 11", "https://company11.com")

	// Should now show pagination with next button enabled
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 11 results")).ToHaveCount(1))
	require.NoError(t, expect.Locator(nextButton).ToBeEnabled())

	// Navigate to page 2 to see there's one more job
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Showing 11 to 11 of 11 results")).ToHaveCount(1))

	// The new job should be visible somewhere (either page 1 or page 2)
	// Let's just verify the count is correct rather than checking specific job location
}

func TestPagination_ArchiveUpdatesCount(t *testing.T) {
	beforeEach(t)
	signin(t, "pagination_user9@email.com", "password")

	// Add 12 jobs to have pagination
	for i := 1; i <= 12; i++ {
		addJobApplication(t, fmt.Sprintf("Company %d", i), fmt.Sprintf("Role %d", i), fmt.Sprintf("https://company%d.com", i))
	}

	// Verify initial state
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 12 results")).ToHaveCount(1))

	// Navigate to page 2
	nextButton := page.Locator("#pagination button").Filter(playwright.LocatorFilterOptions{HasText: "Next"})
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	// Should show 2 items on page 2
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(2))
	require.NoError(t, expect.Locator(page.GetByText("Showing 11 to 12 of 12 results")).ToHaveCount(1))

	// Archive one job from page 2 (whatever job is there)
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	waitForHTMXRequest(t)
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"}).Click())
	waitForHTMXRequest(t)

	// After archiving, we should be redirected back to page 1 with 11 total jobs
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 11 results")).ToHaveCount(1))

	// Navigate back to page 2 to archive the last job
	require.NoError(t, nextButton.Click())
	waitForHTMXRequest(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	waitForHTMXRequest(t)
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"}).Click())
	waitForHTMXRequest(t)

	// Should now be back to page 1 with exactly 10 jobs (no pagination needed)
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(10))
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 10 of 10 results")).ToHaveCount(1))

	// Next button should now be disabled
	require.NoError(t, expect.Locator(nextButton).ToBeDisabled())
}
