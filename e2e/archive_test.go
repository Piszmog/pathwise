//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestArchive_InputFieldsDisabled(t *testing.T) {
	beforeEach(t)
	signin(t, "user10@email.com", "password")

	addJobApplication(t, "Test Company", "Software Engineer", "https://test.com")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))

	archiveSingleJob(t, "Test Company")

	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Test Company")).ToHaveCount(1))

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	require.NoError(t, expect.Locator(page.Locator("#job-form #company")).ToBeDisabled())
	require.NoError(t, expect.Locator(page.Locator("#job-form #title")).ToBeDisabled())
	require.NoError(t, expect.Locator(page.Locator("#job-form #url")).ToBeDisabled())
	require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToBeDisabled())

	salaryMinField := page.Locator("#job-form #salary-min")
	if count, _ := salaryMinField.Count(); count > 0 {
		require.NoError(t, expect.Locator(salaryMinField).ToBeDisabled())
	}

	salaryMaxField := page.Locator("#job-form #salary-max")
	if count, _ := salaryMaxField.Count(); count > 0 {
		require.NoError(t, expect.Locator(salaryMaxField).ToBeDisabled())
	}

	salaryCurrencyField := page.Locator("#job-form #salary-currency")
	if count, _ := salaryCurrencyField.Count(); count > 0 {
		require.NoError(t, expect.Locator(salaryCurrencyField).ToBeDisabled())
	}

	updateButton := page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"})
	require.NoError(t, expect.Locator(updateButton).ToHaveCount(0))

	unarchiveButton := page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Unarchive"})
	require.NoError(t, expect.Locator(unarchiveButton).ToHaveCount(1))
}

func TestArchive_UnarchiveJobApplication(t *testing.T) {
	beforeEach(t)
	signin(t, "user11@email.com", "password")

	addJobApplication(t, "Unarchive Test Company", "Backend Developer", "https://unarchivetest.com")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))

	archiveSingleJob(t, "Unarchive Test Company")

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("0 results")).ToHaveCount(1))

	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Unarchive Test Company")).ToHaveCount(1))

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Unarchive"}).Click())

	// Wait for unarchive operation to complete
	waitForHTMXRequest(t)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("0 results")).ToHaveCount(1))

	_, err = page.Goto(getFullPath(""))
	require.NoError(t, err)

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Unarchive Test Company")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("1 result")).ToHaveCount(1))

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	require.NoError(t, expect.Locator(page.Locator("#job-form #company")).ToBeEnabled())
	require.NoError(t, expect.Locator(page.Locator("#job-form #title")).ToBeEnabled())
	require.NoError(t, expect.Locator(page.Locator("#job-form #url")).ToBeEnabled())
	require.NoError(t, expect.Locator(page.Locator("#job-form #status-select")).ToBeEnabled())

	updateButton := page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"})
	require.NoError(t, expect.Locator(updateButton).ToHaveCount(1))

	archiveButton := page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"})
	require.NoError(t, expect.Locator(archiveButton).ToHaveCount(1))

	unarchiveButton := page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Unarchive"})
	require.NoError(t, expect.Locator(unarchiveButton).ToHaveCount(0))
}
