//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSecurity_UnauthenticatedAccess(t *testing.T) {
	beforeEach(t)

	// Test that unauthenticated users are redirected to signin
	protectedPaths := []string{
		"",
		"archives",
		"settings",
		"export",
	}

	for _, path := range protectedPaths {
		_, err := page.Goto(getFullPath(path))
		require.NoError(t, err)

		// Should be redirected to signin page
		require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
			Timeout: playwright.Float(5000),
		}))
	}
}

func TestSecurity_SessionExpiration(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "security")
	signin(t, user.Email, "password")

	// Verify user is authenticated
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// Clear all cookies to simulate session expiration
	require.NoError(t, context.ClearCookies())

	// Try to access protected page
	_, err := page.Goto(getFullPath(""))
	require.NoError(t, err)

	// Should be redirected to signin
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))
}

func TestSecurity_InvalidCredentials(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "security")

	// Test with wrong password
	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill(user.Email))
	require.NoError(t, page.Locator("#password").Fill("wrongpassword"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	// Should remain on signin page with error
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))
	require.NoError(t, expect.Locator(page.GetByText("Invalid email or password")).ToBeVisible())

	// Test with non-existent email
	require.NoError(t, page.Locator("#email").Fill("nonexistent@example.com"))
	require.NoError(t, page.Locator("#password").Fill("password"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	// Should remain on signin page with error
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))
	require.NoError(t, expect.Locator(page.GetByText("Invalid email or password")).ToBeVisible())
}

func TestSecurity_UserDataIsolation(t *testing.T) {
	beforeEach(t)
	user1 := createTestUser(t, "security1")
	user2 := createTestUser(t, "security2")

	// User 1 creates a job application
	signin(t, user1.Email, "password")
	addJobApplication(t, "User1 Company", "User1 Job", "https://user1.com")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("User1 Company")).ToBeVisible())

	// Sign out user 1
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign out"}).Click())
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// User 2 signs in and should not see user 1's data
	signin(t, user2.Email, "password")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.GetByText("User1 Company")).ToHaveCount(0))

	// User 2 creates their own job application
	addJobApplication(t, "User2 Company", "User2 Job", "https://user2.com")
	require.NoError(t, expect.Locator(page.Locator("#job-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.GetByText("User2 Company")).ToBeVisible())
	require.NoError(t, expect.Locator(page.GetByText("User1 Company")).ToHaveCount(0))
}

func TestSecurity_PasswordValidation(t *testing.T) {
	beforeEach(t)

	_, err := page.Goto(getFullPath("signup"))
	require.NoError(t, err)

	// Test weak password
	require.NoError(t, page.Locator("#email").Fill("test@example.com"))
	require.NoError(t, page.Locator("#password").Fill("123"))
	require.NoError(t, page.Locator("#confirm-password").Fill("123"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	// Should show validation error
	require.NoError(t, expect.Locator(page.GetByText("Password must be at least 8 characters")).ToBeVisible())

	// Test password mismatch
	require.NoError(t, page.Locator("#password").Fill("validpassword123"))
	require.NoError(t, page.Locator("#confirm-password").Fill("differentpassword"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	// Should show mismatch error
	require.NoError(t, expect.Locator(page.GetByText("Passwords do not match")).ToBeVisible())
}

func TestSecurity_SessionPersistence(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "security")
	signin(t, user.Email, "password")

	// Verify user is authenticated
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// Refresh the page
	_, err := page.Reload()
	require.NoError(t, err)

	// Should still be authenticated (not redirected to signin)
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// Navigate to different pages and verify session persists
	_, err = page.Goto(getFullPath("settings"))
	require.NoError(t, err)
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("settings"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	_, err = page.Goto(getFullPath("archives"))
	require.NoError(t, err)
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("archives"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))
}

func TestSecurity_MultipleSessionsSignout(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "security")

	// Create a second browser context to simulate multiple sessions
	context2, page2 := newBrowserContextAndPage(t, defaultContextOptions)
	defer func() { _ = context2.Close() }()

	// Sign in on both sessions
	signin(t, user.Email, "password")

	_, err := page2.Goto(getFullPath("signin"))
	require.NoError(t, err)
	require.NoError(t, page2.Locator("#email").Fill(user.Email))
	require.NoError(t, page2.Locator("#password").Fill("password"))
	require.NoError(t, page2.Locator("button[type=submit]").Click())
	require.NoError(t, expect.Page(page2).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// Both sessions should be active
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))
	require.NoError(t, expect.Page(page2).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// Sign out from first session
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign out"}).Click())
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// Second session should still be active
	_, err = page2.Reload()
	require.NoError(t, err)
	require.NoError(t, expect.Page(page2).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))
}

func TestSecurity_DirectURLAccess(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "security")

	// Try to access protected URLs directly without authentication
	protectedURLs := []string{
		getFullPath(""),
		getFullPath("archives"),
		getFullPath("settings"),
		getFullPath("export"),
	}

	for _, url := range protectedURLs {
		_, err := page.Goto(url)
		require.NoError(t, err)

		// Should be redirected to signin
		require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
			Timeout: playwright.Float(5000),
		}))
	}

	// Sign in
	signin(t, user.Email, "password")

	// Now should be able to access protected URLs
	for _, url := range protectedURLs {
		_, err := page.Goto(url)
		require.NoError(t, err)

		// Should not be redirected to signin
		require.NoError(t, expect.Page(page).Not().ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
			Timeout: playwright.Float(2000),
		}))
	}
}

func TestSecurity_FormValidationRequired(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "security")
	signin(t, user.Email, "password")

	// Test job application form validation
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())

	// Try to submit empty form
	require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

	// Form should not submit (HTML5 validation should prevent it)
	// The form should still be visible
	require.NoError(t, expect.Locator(page.Locator("#new-job-form")).ToBeVisible())

	// Fill required fields and submit should work
	require.NoError(t, page.Locator("#new-job-form #company").Fill("Test Company"))
	require.NoError(t, page.Locator("#new-job-form #title").Fill("Test Job"))
	require.NoError(t, page.Locator("#new-job-form #url").Fill("https://test.com"))
	require.NoError(t, page.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

	// Should successfully add the job
	require.NoError(t, expect.Locator(page.GetByText("Test Company")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))
}

func TestSecurity_ConcurrentUserOperations(t *testing.T) {
	beforeEach(t)
	user1 := createTestUser(t, "concurrent1")
	user2 := createTestUser(t, "concurrent2")

	// Create a second browser context for user2
	context2, page2 := newBrowserContextAndPage(t, defaultContextOptions)
	defer func() { _ = context2.Close() }()

	// Both users sign in simultaneously
	signin(t, user1.Email, "password")

	_, err := page2.Goto(getFullPath("signin"))
	require.NoError(t, err)
	require.NoError(t, page2.Locator("#email").Fill(user2.Email))
	require.NoError(t, page2.Locator("#password").Fill("password"))
	require.NoError(t, page2.Locator("button[type=submit]").Click())
	require.NoError(t, expect.Page(page2).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// Both users add job applications simultaneously
	addJobApplication(t, "User1 Company", "User1 Job", "https://user1.com")

	require.NoError(t, page2.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Add"}).First().Click())
	require.NoError(t, page2.Locator("#new-job-form #company").Fill("User2 Company"))
	require.NoError(t, page2.Locator("#new-job-form #title").Fill("User2 Job"))
	require.NoError(t, page2.Locator("#new-job-form #url").Fill("https://user2.com"))
	require.NoError(t, page2.Locator("#new-job-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

	// Wait for both operations to complete
	require.NoError(t, expect.Locator(page.GetByText("User1 Company")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))
	require.NoError(t, expect.Locator(page2.GetByText("User2 Company")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Verify data isolation - each user should only see their own data
	require.NoError(t, expect.Locator(page.GetByText("User2 Company")).ToHaveCount(0))
	require.NoError(t, expect.Locator(page2.GetByText("User1 Company")).ToHaveCount(0))

	// Both users should have correct stats
	require.NoError(t, expect.Locator(page.Locator("#stats-total-applications")).ToHaveText("Total Applications1"))
	require.NoError(t, expect.Locator(page2.Locator("#stats-total-applications")).ToHaveText("Total Applications1"))
}

func TestSecurity_SignoutInvalidatesSession(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "security")
	signin(t, user.Email, "password")

	// Verify authenticated access
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// Sign out
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign out"}).Click())
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// Try to access protected page using browser back button
	_, err := page.GoBack()
	require.NoError(t, err)

	// Should be redirected back to signin (session invalidated)
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	// Try direct URL access after signout
	_, err = page.Goto(getFullPath(""))
	require.NoError(t, err)
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))
}
