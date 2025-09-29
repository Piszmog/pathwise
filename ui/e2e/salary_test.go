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

	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToHaveValue("GBP", playwright.LocatorAssertionsToHaveValueOptions{
		Timeout: playwright.Float(5000),
	}))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_min")).ToHaveValue("45000"))
	require.NoError(t, expect.Locator(page.Locator("#job-form #salary_max")).ToHaveValue("65000"))
}

func TestSalary_ValidationScenarios(t *testing.T) {
	testCases := []struct {
		name       string
		company    string
		title      string
		url        string
		currency   string
		minSalary  string
		maxSalary  string
		shouldFail bool
		isUpdate   bool
	}{
		{
			name:       "min greater than max",
			company:    "Validation Test Company",
			title:      "DevOps Engineer",
			url:        "https://validationtest.com",
			currency:   "USD",
			minSalary:  "120000",
			maxSalary:  "80000",
			shouldFail: true,
			isUpdate:   false,
		},
		{
			name:       "negative salary values",
			company:    "Invalid Salary Company",
			title:      "QA Engineer",
			url:        "https://invalidsalary.com",
			currency:   "USD",
			minSalary:  "-50000",
			maxSalary:  "-30000",
			shouldFail: false,
			isUpdate:   true,
		},
		{
			name:       "extremely large values",
			company:    "Large Salary Company",
			title:      "Senior Engineer",
			url:        "https://largesalary.com",
			currency:   "USD",
			minSalary:  "999999999",
			maxSalary:  "9999999999",
			shouldFail: false,
			isUpdate:   false,
		},
		{
			name:       "zero values",
			company:    "Zero Salary Company",
			title:      "Intern",
			url:        "https://zerosalary.com",
			currency:   "USD",
			minSalary:  "0",
			maxSalary:  "0",
			shouldFail: false,
			isUpdate:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			beforeEach(t)
			createUserAndSignIn(t)

			if tc.isUpdate {
				addJobApplication(t, tc.company, tc.title, tc.url)

				require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

				require.NoError(t, expect.Locator(page.Locator("#job-form")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
					Timeout: playwright.Float(5000),
				}))

				if tc.currency != "" {
					_, err := page.Locator("#job-form #salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{tc.currency}})
					require.NoError(t, err)
				}
				require.NoError(t, page.Locator("#job-form #salary_min").Fill(tc.minSalary))
				require.NoError(t, page.Locator("#job-form #salary_max").Fill(tc.maxSalary))

				require.NoError(t, page.Locator("#job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Update"}).Click())
			} else {
				require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
				require.NoError(t, page.Locator("#new-job-form #company").Fill(tc.company))
				require.NoError(t, page.Locator("#new-job-form #title").Fill(tc.title))
				require.NoError(t, page.Locator("#new-job-form #url").Fill(tc.url))

				if tc.currency != "" {
					_, err := page.Locator("#new-job-form #new-salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{tc.currency}})
					require.NoError(t, err)
				}
				require.NoError(t, page.Locator("#new-job-form #new-salary_min").Fill(tc.minSalary))
				require.NoError(t, page.Locator("#new-job-form #new-salary_max").Fill(tc.maxSalary))

				require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

				if !tc.shouldFail {
					require.NoError(t, expect.Locator(page.GetByText(tc.company)).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
						Timeout: playwright.Float(10000),
					}))
				}
			}
		})
	}
}

func TestSalary_PartialSalaryInformation(t *testing.T) {
	testCases := []struct {
		name             string
		company          string
		title            string
		url              string
		currency         string
		minSalary        string
		maxSalary        string
		expectedCurrency string
		expectedMin      string
		expectedMax      string
	}{
		{
			name:             "minimum salary only",
			company:          "Min Only Company",
			title:            "Data Scientist",
			url:              "https://minonly.com",
			currency:         "CAD",
			minSalary:        "70000",
			maxSalary:        "",
			expectedCurrency: "CAD",
			expectedMin:      "70000",
			expectedMax:      "",
		},
		{
			name:             "maximum salary only",
			company:          "Max Only Company",
			title:            "Product Manager",
			url:              "https://maxonly.com",
			currency:         "USD",
			minSalary:        "",
			maxSalary:        "150000",
			expectedCurrency: "USD",
			expectedMin:      "",
			expectedMax:      "150000",
		},
		{
			name:             "currency only",
			company:          "Currency Only Company",
			title:            "Designer",
			url:              "https://currencyonly.com",
			currency:         "EUR",
			minSalary:        "",
			maxSalary:        "",
			expectedCurrency: "EUR",
			expectedMin:      "",
			expectedMax:      "",
		},
		{
			name:             "no salary information",
			company:          "No Salary Company",
			title:            "Intern",
			url:              "https://nosalary.com",
			currency:         "",
			minSalary:        "",
			maxSalary:        "",
			expectedCurrency: "",
			expectedMin:      "",
			expectedMax:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			beforeEach(t)
			createUserAndSignIn(t)

			require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
			require.NoError(t, page.Locator("#new-job-form #company").Fill(tc.company))
			require.NoError(t, page.Locator("#new-job-form #title").Fill(tc.title))
			require.NoError(t, page.Locator("#new-job-form #url").Fill(tc.url))

			if tc.currency != "" {
				_, err := page.Locator("#new-job-form #new-salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{tc.currency}})
				require.NoError(t, err)
			}
			if tc.minSalary != "" {
				require.NoError(t, page.Locator("#new-job-form #new-salary_min").Fill(tc.minSalary))
			}
			if tc.maxSalary != "" {
				require.NoError(t, page.Locator("#new-job-form #new-salary_max").Fill(tc.maxSalary))
			}

			require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

			require.NoError(t, expect.Locator(page.GetByText(tc.company)).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
				Timeout: playwright.Float(10000),
			}))

			require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

			require.NoError(t, expect.Locator(page.Locator("#job-form")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
				Timeout: playwright.Float(5000),
			}))

			require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToHaveValue(tc.expectedCurrency))
			require.NoError(t, expect.Locator(page.Locator("#job-form #salary_min")).ToHaveValue(tc.expectedMin))
			require.NoError(t, expect.Locator(page.Locator("#job-form #salary_max")).ToHaveValue(tc.expectedMax))
		})
	}
}

func TestSalary_AllCurrencyOptions(t *testing.T) {
	testCases := []struct {
		name     string
		currency string
		minSal   string
		maxSal   string
	}{
		{"USD currency", "USD", "50000", "80000"},
		{"EUR currency", "EUR", "45000", "70000"},
		{"GBP currency", "GBP", "40000", "65000"},
		{"CAD currency", "CAD", "55000", "85000"},
		{"AUD currency", "AUD", "60000", "90000"},
		{"JPY currency", "JPY", "5000000", "8000000"},
		{"CHF currency", "CHF", "48000", "75000"},
	}

	beforeEach(t)
	createUserAndSignIn(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			companyName := "Currency Test " + tc.currency

			require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
			require.NoError(t, page.Locator("#new-job-form #company").Fill(companyName))
			require.NoError(t, page.Locator("#new-job-form #title").Fill("Software Engineer"))
			require.NoError(t, page.Locator("#new-job-form #url").Fill("https://currencytest"+tc.currency+".com"))

			_, err := page.Locator("#new-job-form #new-salary_currency").SelectOption(playwright.SelectOptionValues{Values: &[]string{tc.currency}})
			require.NoError(t, err)
			require.NoError(t, page.Locator("#new-job-form #new-salary_min").Fill(tc.minSal))
			require.NoError(t, page.Locator("#new-job-form #new-salary_max").Fill(tc.maxSal))

			require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

			require.NoError(t, expect.Locator(page.GetByText(companyName)).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
				Timeout: playwright.Float(10000),
			}))

			jobRow := page.GetByText(companyName).Locator("xpath=ancestor::li")
			require.NoError(t, jobRow.GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "View job"}).Click())

			require.NoError(t, expect.Locator(page.Locator("#job-form")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
				Timeout: playwright.Float(5000),
			}))

			require.NoError(t, expect.Locator(page.Locator("#job-form #salary_currency")).ToHaveValue(tc.currency))
			require.NoError(t, expect.Locator(page.Locator("#job-form #salary_min")).ToHaveValue(tc.minSal))
			require.NoError(t, expect.Locator(page.Locator("#job-form #salary_max")).ToHaveValue(tc.maxSal))

			require.NoError(t, page.Locator("body").Click())
		})
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
