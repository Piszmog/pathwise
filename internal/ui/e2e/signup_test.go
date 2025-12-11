//go:build e2e

package e2e_test

import (
	"testing"
	"time"

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
	beforeEach(t)
	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("a[href='/signup']").Click())

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signup"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSignUp_UserAlreadyExists(t *testing.T) {
	beforeEach(t)
	_, err := page.Goto(getFullPath("signup"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("duplicate@email.com"))
	require.NoError(t, page.Locator("#password").Fill("MySuperPassword1234!"))
	require.NoError(t, page.Locator("#confirmPassword").Fill("MySuperPassword1234!"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))

	_, err = page.Goto(getFullPath("signup"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("duplicate@email.com"))
	require.NoError(t, page.Locator("#password").Fill("MySuperPassword5678!"))
	require.NoError(t, page.Locator("#confirmPassword").Fill("MySuperPassword5678!"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	require.NoError(t, expect.Locator(page.GetByText("Something went wrong")).ToBeHidden())
}

func TestSignUp_InvalidEmail(t *testing.T) {
	beforeEach(t)
	_, err := page.Goto(getFullPath("signup"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("invalid-email"))
	require.NoError(t, page.Locator("#password").Fill("MySuperPassword1234!"))
	require.NoError(t, page.Locator("#confirmPassword").Fill("MySuperPassword1234!"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signup"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSignUp_InvalidPassword(t *testing.T) {
	beforeEach(t)
	_, err := page.Goto(getFullPath("signup"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("invalidpass@email.com"))
	require.NoError(t, page.Locator("#password").Fill("password"))
	require.NoError(t, page.Locator("#confirmPassword").Fill("password"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signup"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSignUp_PasswordsDoNotMatch(t *testing.T) {
	beforeEach(t)
	_, err := page.Goto(getFullPath("signup"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill("invalidpass@email.com"))
	require.NoError(t, page.Locator("#password").Fill("MySuperPassword1234!"))
	require.NoError(t, page.Locator("#confirmPassword").Fill("MySuperPassword1234"))
	require.NoError(t, page.Locator("button[type=submit]").Click())

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signup"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(10000),
	}))
}

func TestSignup_ClearsCookiesOnPageLoad(t *testing.T) {
	beforeEach(t)
	email := generateUniqueEmail(t)
	createTestUser(t, email)

	signin(t, email, "password")

	cookies, err := ctx.Cookies()
	require.NoError(t, err)
	require.NotEmpty(t, cookies)

	sessionCookie := findCookie(cookies, "session")
	require.NotNil(t, sessionCookie, "session cookie should exist after signin")

	_, err = page.Goto(getFullPath("signup"))
	require.NoError(t, err)

	cookies, err = ctx.Cookies()
	require.NoError(t, err)

	sessionCookie = findCookie(cookies, "session")
	if sessionCookie != nil {
		require.True(t, time.Unix(int64(sessionCookie.Expires), 0).Before(time.Now()),
			"session cookie should be expired after visiting signup page")
	}
}
