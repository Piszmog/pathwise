//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestHome_NewUser(t *testing.T) {
	beforeEach(t)
	signin(t)

	// Initial stats
	assertStats(t, "0", "0", "0 days", "NaN%", "NaN%")

	// Initial job apps - empty
	require.NoError(t, expect.Locator(page.Locator("#job-list").GetByRole("li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("Showing 0 to 0 of 0 results ")).ToHaveCount(1))
}

func TestHome_AddApplication(t *testing.T) {
	beforeEach(t)
	signin(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list").GetByRole("li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("Showing 0 to 0 of 0 results")).ToHaveCount(1))

	addJobApplication(t, "Super Company", "Rock Star", "https://supercompany.com")
	require.NoError(t, expect.Locator(page.Locator("#job-1-row")).ToBeVisible())
	require.NoError(t, expect.Locator(page.GetByText("Showing 1 to 1 of 1 results")).ToHaveCount(1))

	assertStats(t, "1", "1", "0 days", "0%", "0%")
}

func signin(t *testing.T) {
	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("user1@email.com"))
	require.NoError(t, page.Locator("#password").Fill("password"))
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
