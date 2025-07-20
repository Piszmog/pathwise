//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSalary_AddJobWithSalaryFields(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job with all salary fields
	addJobApplicationWithSalary(t, "Salary Test Company", "Software Engineer", "https://salarytest.com", "80000", "120000", "USD")

	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("Salary Test Company")).ToHaveCount(1))

	// View job details to verify salary fields were saved
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Verify salary fields are populated correctly
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue("USD"))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue("80000"))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue("120000"))
}

func TestSalary_AddJobWithPartialSalaryFields(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job with only min salary
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
	require.NoError(t, page.Locator("#new-job-form #company").Fill("Partial Salary Company"))
	require.NoError(t, page.Locator("#new-job-form #title").Fill("Backend Developer"))
	require.NoError(t, page.Locator("#new-job-form #url").Fill("https://partialsalary.com"))
	require.NoError(t, page.Locator("#new-job-form #new-salary_min").Fill("90000"))
	// Leave max salary and currency empty
	require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

	require.NoError(t, expect.Locator(page.GetByText("Partial Salary Company")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(10000),
	}))

	// View job details to verify partial salary fields
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue("90000"))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue(""))
}

func TestSalary_AddJobWithNoSalaryFields(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job without any salary information
	addJobApplication(t, "No Salary Company", "Frontend Developer", "https://nosalary.com")

	// View job details to verify empty salary fields
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue(""))
}

func TestSalary_UpdateExistingSalaryFields(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job with initial salary
	addJobApplicationWithSalary(t, "Update Salary Company", "DevOps Engineer", "https://updatesalary.com", "70000", "100000", "EUR")

	// View job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Verify initial values
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue("EUR"))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue("70000"))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue("100000"))

	// Update salary fields
	_, err := page.Locator("#salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{"USD"}})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#salary_min").Fill("85000"))
	require.NoError(t, page.Locator("#salary_max").Fill("130000"))

	// Submit update
	require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())

	// Wait for update to complete
	waitForHTMXRequest(t)

	// Verify updated values
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue("USD"))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue("85000"))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue("130000"))
}

func TestSalary_ClearSalaryFields(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job with salary
	addJobApplicationWithSalary(t, "Clear Salary Company", "Product Manager", "https://clearsalary.com", "95000", "140000", "GBP")

	// View job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Clear all salary fields
	_, err := page.Locator("#salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{""}})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#salary_min").Fill(""))
	require.NoError(t, page.Locator("#salary_max").Fill(""))

	// Submit update
	require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())

	// Wait for update to complete
	waitForHTMXRequest(t)

	// Verify fields are cleared
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue(""))
}

func TestSalary_AllCurrencyOptions(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job to test currency options
	addJobApplication(t, "Currency Test Company", "Data Scientist", "https://currencytest.com")

	// View job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Test each available currency
	currencies := []string{"USD", "EUR", "GBP", "CAD", "AUD", "JPY", "CHF"}

	for _, currency := range currencies {
		_, err := page.Locator("#salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{currency}})
		require.NoError(t, err)
		require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue(currency))
	}

	// Test setting back to empty
	_, err := page.Locator("#salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{""}})
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue(""))
}

func TestSalary_NumericValidation(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job to test numeric validation
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
	require.NoError(t, page.Locator("#new-job-form #company").Fill("Validation Test Company"))
	require.NoError(t, page.Locator("#new-job-form #title").Fill("QA Engineer"))
	require.NoError(t, page.Locator("#new-job-form #url").Fill("https://validationtest.com"))

	// Test with valid numbers directly (Playwright doesn't allow typing non-numeric into number inputs)
	require.NoError(t, page.Locator("#new-job-form #new-salary_min").Fill("50000"))
	require.NoError(t, page.Locator("#new-job-form #new-salary_max").Fill("75000"))

	require.NoError(t, expect.Locator(page.Locator("#new-job-form #new-salary_min")).ToHaveValue("50000"))
	require.NoError(t, expect.Locator(page.Locator("#new-job-form #new-salary_max")).ToHaveValue("75000"))

	// Submit the form
	require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

	require.NoError(t, expect.Locator(page.GetByText("Validation Test Company")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSalary_LargeNumbers(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Test with large salary numbers
	addJobApplicationWithSalary(t, "High Salary Company", "Senior Architect", "https://highsalary.com", "250000", "500000", "USD")

	// View job details to verify large numbers are handled correctly
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue("USD"))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue("250000"))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue("500000"))
}

func TestSalary_ZeroValues(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Test with zero values
	addJobApplicationWithSalary(t, "Zero Salary Company", "Intern", "https://zerosalary.com", "0", "0", "USD")

	// View job details to verify zero values are handled correctly
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue("USD"))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue("0"))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue("0"))
}

func TestSalary_ArchivedJobSalaryFieldsDisabled(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job with salary
	addJobApplicationWithSalary(t, "Archive Salary Company", "Solutions Architect", "https://archivesalary.com", "120000", "180000", "CAD")

	// Archive the job
	archiveSingleJob(t, "Archive Salary Company")

	// Go to archives page
	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)

	// View archived job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Verify salary fields are disabled
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToBeDisabled())
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToBeDisabled())
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToBeDisabled())

	// Verify values are still displayed correctly
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue("CAD"))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue("120000"))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue("180000"))
}

func TestSalary_UnarchiveJobSalaryFieldsEnabled(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job with salary
	addJobApplicationWithSalary(t, "Unarchive Salary Company", "Tech Lead", "https://unarchivesalary.com", "140000", "200000", "AUD")

	// Archive the job
	archiveSingleJob(t, "Unarchive Salary Company")

	// Go to archives page and unarchive
	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Unarchive"}).Click())

	// Wait for unarchive operation
	waitForHTMXRequest(t)

	// Go back to home page
	_, err = page.Goto(getFullPath(""))
	require.NoError(t, err)

	// View unarchived job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Verify salary fields are enabled again
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToBeEnabled())
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToBeEnabled())
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToBeEnabled())

	// Verify values are preserved
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue("AUD"))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue("140000"))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue("200000"))
}

func TestSalary_UpdateOnlyMinSalary(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job without salary
	addJobApplication(t, "Min Only Company", "Software Engineer", "https://minonly.com")

	// View job details and add only min salary
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	require.NoError(t, page.Locator("#salary_min").Fill("60000"))
	// Leave max and currency empty

	// Submit update
	require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())

	// Wait for update
	waitForHTMXRequest(t)

	// Verify only min salary is set
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue("60000"))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue(""))
}

func TestSalary_UpdateOnlyMaxSalary(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job without salary
	addJobApplication(t, "Max Only Company", "Product Manager", "https://maxonly.com")

	// View job details and add only max salary
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	require.NoError(t, page.Locator("#salary_max").Fill("150000"))
	// Leave min and currency empty

	// Submit update
	require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())

	// Wait for update
	waitForHTMXRequest(t)

	// Verify only max salary is set
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue("150000"))
}

func TestSalary_UpdateOnlyCurrency(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "salary")
	signin(t, user.Email, "password")

	// Add job without salary
	addJobApplication(t, "Currency Only Company", "Data Analyst", "https://currencyonly.com")

	// View job details and add only currency
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	_, err := page.Locator("#salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{"JPY"}})
	require.NoError(t, err)
	// Leave min and max empty

	// Submit update
	require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())

	// Wait for update
	waitForHTMXRequest(t)

	// Verify only currency is set
	require.NoError(t, expect.Locator(page.Locator("#salary_currency")).ToHaveValue("JPY"))
	require.NoError(t, expect.Locator(page.Locator("#salary_min")).ToHaveValue(""))
	require.NoError(t, expect.Locator(page.Locator("#salary_max")).ToHaveValue(""))
}
