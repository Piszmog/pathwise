//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSignin(t *testing.T) {
	beforeEach(t)
	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("user1@email.com"))
	require.NoError(t, page.Locator("#password").Fill("password"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSignin_InvalidCredentials(t *testing.T) {
	beforeEach(t)
	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("user_does_not_exist@email.com"))
	require.NoError(t, page.Locator("#password").Fill("password"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	require.NoError(t, expect.Locator(page.GetByText("Incorrect email or password")).ToBeVisible())
}

func TestSignin_AuthMiddleware(t *testing.T) {
	beforeEach(t)
	_, err := page.Goto(getFullPath("") + "/")
	require.NoError(t, err)

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}
