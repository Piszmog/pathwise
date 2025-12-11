//go:build e2e

package e2e_test

import (
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSignin(t *testing.T) {
	beforeEach(t)
	email := generateUniqueEmail(t)
	createTestUser(t, email)

	_, err := page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	require.NoError(t, page.Locator("#email").Fill(email))
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

func TestSignin_ClearsCookiesOnPageLoad(t *testing.T) {
	beforeEach(t)
	email := generateUniqueEmail(t)
	createTestUser(t, email)

	signin(t, email, "password")

	cookies, err := ctx.Cookies()
	require.NoError(t, err)
	require.NotEmpty(t, cookies)

	sessionCookie := findCookie(cookies, "session")
	require.NotNil(t, sessionCookie, "session cookie should exist after signin")

	_, err = page.Goto(getFullPath("signin"))
	require.NoError(t, err)

	cookies, err = ctx.Cookies()
	require.NoError(t, err)

	sessionCookie = findCookie(cookies, "session")
	if sessionCookie != nil {
		require.True(t, time.Unix(int64(sessionCookie.Expires), 0).Before(time.Now()),
			"session cookie should be expired after visiting signin page")
	}
}

func findCookie(cookies []playwright.Cookie, name string) *playwright.Cookie {
	for i, cookie := range cookies {
		if cookie.Name == name {
			return &cookies[i]
		}
	}
	return nil
}
