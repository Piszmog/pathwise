//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSignup(t *testing.T) {
	beforeEach(t)
	_, err := page.Goto(getFullPath("signup"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("test@email.com"))
	require.NoError(t, page.Locator("#password").Fill("MySuperPassword1234!"))
	require.NoError(t, page.Locator("#confirmPassword").Fill("MySuperPassword1234!"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSignUp_NavigationFromSignIn(t *testing.T) {

}
