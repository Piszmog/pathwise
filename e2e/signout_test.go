//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSignout_FromHeader(t *testing.T) {
	beforeEach(t)

	// First sign in
	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("user1@email.com"))
	require.NoError(t, page.Locator("#password").Fill("password"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	// Verify we're on the home page
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))

	// Click the user menu button to open dropdown
	require.NoError(t, page.Locator("button[id='user-menu-button']").Click())

	// Click sign out link in the desktop menu (first one)
	require.NoError(t, page.Locator("a[href='/signout']").First().Click())

	// Verify we're redirected to signin page
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSignout_FromMobileMenu(t *testing.T) {
	beforeEach(t, playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{Width: 375, Height: 667}, // Mobile viewport
	})

	// First sign in
	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("user1@email.com"))
	require.NoError(t, page.Locator("#password").Fill("password"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	// Verify we're on the home page
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))

	// Click the mobile menu button
	require.NoError(t, page.Locator("button[aria-controls='mobile-menu']").Click())

	// Click sign out link in mobile menu
	require.NoError(t, page.Locator("a[href='/signout']").Last().Click())

	// Verify we're redirected to signin page
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSignout_DirectURL(t *testing.T) {
	beforeEach(t)

	// First sign in
	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("user1@email.com"))
	require.NoError(t, page.Locator("#password").Fill("password"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	// Verify we're on the home page
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))

	// Navigate directly to signout URL
	_, err = page.Goto(getFullPath("signout"))
	require.NoError(t, err)

	// Verify we're redirected to signin page
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSignout_SessionInvalidated(t *testing.T) {
	beforeEach(t)

	// First sign in
	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("user1@email.com"))
	require.NoError(t, page.Locator("#password").Fill("password"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	// Verify we're on the home page
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))

	// Sign out
	_, err = page.Goto(getFullPath("signout"))
	require.NoError(t, err)

	// Try to access protected page - should redirect to signin
	_, err = page.Goto(getFullPath("") + "/")
	require.NoError(t, err)

	// Verify we're redirected to signin page (session is invalidated)
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSignout_WithoutSession(t *testing.T) {
	beforeEach(t)

	// Navigate directly to signout URL without signing in first
	_, err := page.Goto(getFullPath("signout"))
	require.NoError(t, err)

	// Should still redirect to signin page (graceful handling of missing session)
	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}
