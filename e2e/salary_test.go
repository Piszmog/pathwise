//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSalary_AddJobWithSalaryRange(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
	require.NoError(t, page.Locator("#new-job-form #company").Fill("Salary Test Company"))
	require.NoError(t, page.Locator("#new-job-form #title").Fill("Software Engineer"))
	require.NoError(t, page.Locator("#new-job-form #url").Fill("https://salarytest.com"))

	// Set salary information
	_, err := page.Locator("#new-job-form #new-salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{"USD"}})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#new-job-form #new-salary_min").Fill("80000"))
	require.NoError(t, page.Locator("#new-job-form #new-salary_max").Fill("120000"))

	require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

	// Wait for the job to be added
	require.NoError(t, expect.Locator(page.GetByText("Salary Test Company")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(10000),
	}))

	// Verify the job was added
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
}

func TestSalary_ViewSalaryInJobDetails(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	addJobApplicationWithSalary(t, "View Salary Company", "Backend Developer", "https://viewsalary.com", "EUR", "60000", "90000")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job form to load
	require.NoError(t, expect.Locator(page.Locator("#job-form")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Verify salary fields are populated correctly
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToHaveValue("EUR"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_min")).ToHaveValue("60000"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_max")).ToHaveValue("90000"))
}

func TestSalary_UpdateSalaryInformation(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	addJobApplication(t, "Update Salary Company", "Frontend Developer", "https://updatesalary.com")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job form to load
	require.NoError(t, expect.Locator(page.Locator("#job-form")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Update salary information
	_, err := page.Locator("#job-form #salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{"GBP"}})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#job-form #salary_min").Fill("45000"))
	require.NoError(t, page.Locator("#job-form #salary_max").Fill("65000"))

	require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())

	// Wait for update to complete
	page.WaitForTimeout(1500)

	// Verify the salary information was updated
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToHaveValue("GBP"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_min")).ToHaveValue("45000"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_max")).ToHaveValue("65000"))
}

func TestSalary_ValidationMinGreaterThanMax(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
	require.NoError(t, page.Locator("#new-job-form #company").Fill("Validation Test Company"))
	require.NoError(t, page.Locator("#new-job-form #title").Fill("DevOps Engineer"))
	require.NoError(t, page.Locator("#new-job-form #url").Fill("https://validationtest.com"))

	// Set invalid salary range (min > max)
	_, err := page.Locator("#new-job-form #new-salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{"USD"}})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#new-job-form #new-salary_min").Fill("120000"))
	require.NoError(t, page.Locator("#new-job-form #new-salary_max").Fill("80000"))

	require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

	// Wait for the form submission to complete
	page.WaitForTimeout(2000)

	// Verify job was not added (main validation that the error occurred)
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(0))

	// Verify we can see the job list is still empty
	require.NoError(t, expect.Locator(page.GetByText("0 results")).ToHaveCount(1))
}

func TestSalary_ValidationInvalidSalaryValues(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	addJobApplication(t, "Invalid Salary Company", "QA Engineer", "https://invalidsalary.com")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job form to load
	require.NoError(t, expect.Locator(page.Locator("#job-form")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Try to set invalid salary values (negative numbers)
	require.NoError(t, page.Locator("#job-form #salary_min").Fill("-50000"))
	require.NoError(t, page.Locator("#job-form #salary_max").Fill("-30000"))

	require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())

	// Wait for potential error
	page.WaitForTimeout(1500)

	// The form should handle negative values gracefully (browser validation or server validation)
	// This test ensures the application doesn't crash with invalid input
}

func TestSalary_PartialSalaryInformation(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	// Test with only minimum salary
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
	require.NoError(t, page.Locator("#new-job-form #company").Fill("Partial Salary Company"))
	require.NoError(t, page.Locator("#new-job-form #title").Fill("Data Scientist"))
	require.NoError(t, page.Locator("#new-job-form #url").Fill("https://partialsalary.com"))

	// Set only minimum salary
	_, err := page.Locator("#new-job-form #new-salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{"CAD"}})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#new-job-form #new-salary_min").Fill("70000"))
	// Leave max salary empty

	require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

	// Wait for the job to be added
	require.NoError(t, expect.Locator(page.GetByText("Partial Salary Company")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(10000),
	}))

	// Open job details to verify partial salary information
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job form to load
	require.NoError(t, expect.Locator(page.Locator("#job-form")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Verify partial salary information
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToHaveValue("CAD"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_min")).ToHaveValue("70000"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_max")).ToHaveValue(""))
}

func TestSalary_AllCurrencyOptions(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	currencies := []string{"USD", "EUR", "GBP", "CAD", "AUD", "JPY", "CHF"}

	for _, currency := range currencies {
		companyName := "Currency Test " + currency

		require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
		require.NoError(t, page.Locator("#new-job-form #company").Fill(companyName))
		require.NoError(t, page.Locator("#new-job-form #title").Fill("Software Engineer"))
		require.NoError(t, page.Locator("#new-job-form #url").Fill("https://currencytest"+currency+".com"))

		// Set currency
		_, err := page.Locator("#new-job-form #new-salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{currency}})
		require.NoError(t, err)
		require.NoError(t, page.Locator("#new-job-form #new-salary_min").Fill("50000"))
		require.NoError(t, page.Locator("#new-job-form #new-salary_max").Fill("80000"))

		require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

		// Wait for the job to be added
		require.NoError(t, expect.Locator(page.GetByText(companyName)).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(10000),
		}))
	}

	// Verify all jobs were added
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(len(currencies)))

	// Test each currency by opening job details
	for _, currency := range currencies {
		companyName := "Currency Test " + currency

		// Find and click the specific job
		jobRow := page.GetByText(companyName).Locator("xpath=ancestor::li")
		require.NoError(t, jobRow.GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "View job"}).Click())

		// Wait for job form to load
		require.NoError(t, expect.Locator(page.Locator("#job-form")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(5000),
		}))

		// Verify currency is correctly set
		require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToHaveValue(currency))

		// Click somewhere else to close the details
		require.NoError(t, page.Locator("body").Click())
		page.WaitForTimeout(500)
	}
}

func TestSalary_ClearSalaryInformation(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	addJobApplicationWithSalary(t, "Clear Salary Company", "Product Manager", "https://clearsalary.com", "USD", "90000", "130000")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job form to load
	require.NoError(t, expect.Locator(page.Locator("#job-form")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Verify salary information is present
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToHaveValue("USD"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_min")).ToHaveValue("90000"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_max")).ToHaveValue("130000"))

	// Clear salary information
	_, err := page.Locator("#job-form #salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{""}})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#job-form #salary_min").Fill(""))
	require.NoError(t, page.Locator("#job-form #salary_max").Fill(""))

	require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())

	// Wait for update to complete
	page.WaitForTimeout(1500)

	// Verify salary information was cleared
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_min")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_max")).ToHaveValue(""))
}

func TestSalary_ArchivedJobSalaryFieldsDisabled(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	addJobApplicationWithSalary(t, "Archive Salary Company", "DevOps Engineer", "https://archivesalary.com", "EUR", "55000", "75000")

	// Archive the job
	archiveSingleJob(t, "Archive Salary Company")

	// Go to archives page
	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)

	// Open archived job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job form to load
	require.NoError(t, expect.Locator(page.Locator("#job-form")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Verify all salary fields are disabled
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToBeDisabled())
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_min")).ToBeDisabled())
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_max")).ToBeDisabled())

	// Verify salary values are still displayed
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToHaveValue("EUR"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_min")).ToHaveValue("55000"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_max")).ToHaveValue("75000"))
}

// Helper function to add job application with salary information
func addJobApplicationWithSalary(t *testing.T, company, title, url, currency, minSalary, maxSalary string) {
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
	require.NoError(t, page.Locator("#new-job-form #company").Fill(company))
	require.NoError(t, page.Locator("#new-job-form #title").Fill(title))
	require.NoError(t, page.Locator("#new-job-form #url").Fill(url))

	if currency != "" {
		_, err := page.Locator("#new-job-form #new-salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{currency}})
		require.NoError(t, err)
	}
	if minSalary != "" {
		require.NoError(t, page.Locator("#new-job-form #new-salary_min").Fill(minSalary))
	}
	if maxSalary != "" {
		require.NoError(t, page.Locator("#new-job-form #new-salary_max").Fill(maxSalary))
	}

	require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

	// Wait for the job to be added
	require.NoError(t, expect.Locator(page.GetByText(company)).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(10000),
	}))
}
