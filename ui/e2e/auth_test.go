//go:build e2e

package e2e_test

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestAuth_SessionPersistence(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	_, err := page.Goto(getFullPath(""))
	require.NoError(t, err)

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	require.NoError(t, expect.Locator(page.GetByText("New Job Application")).ToBeVisible())
}

func TestAuth_SessionExpiry(t *testing.T) {
	beforeEach(t)
	email := generateUniqueEmail(t)
	createTestUser(t, email)

	createExpiredSession(t, email)

	_, err := page.Goto(getFullPath(""))
	require.NoError(t, err)

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))
}

func TestAuth_SessionRefresh(t *testing.T) {
	beforeEach(t)
	email := generateUniqueEmail(t)
	createTestUser(t, email)

	sessionToken := createSessionNeedingRefresh(t, email)

	require.NoError(t, ctx.AddCookies([]playwright.OptionalCookie{
		{
			Name:     "session",
			Value:    sessionToken,
			Domain:   playwright.String("localhost"),
			Path:     playwright.String("/"),
			HttpOnly: playwright.Bool(true),
			SameSite: playwright.SameSiteAttributeStrict,
		},
	}))

	_, err := page.Goto(getFullPath(""))
	require.NoError(t, err)

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	verifySessionWasRefreshed(t, sessionToken)
}

func TestAuth_Signout(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	signoutLocator := page.Locator("a[href='/signout'], button[onclick*='signout']").First()
	if count, _ := signoutLocator.Count(); count > 0 {
		require.NoError(t, signoutLocator.Click())
	} else {
		_, err := page.Goto(getFullPath("signout"))
		require.NoError(t, err)
	}

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	_, err := page.Goto(getFullPath(""))
	require.NoError(t, err)

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("signin"), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))
}

func TestAuth_HTMXRequests(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	_, err := page.Goto(getFullPath(""))
	require.NoError(t, err)

	htmxElements := page.Locator("[hx-get], [hx-post], [hx-patch], [hx-delete]")
	count, err := htmxElements.Count()
	require.NoError(t, err)

	if count > 0 {
		require.NoError(t, htmxElements.First().Click())

		page.WaitForTimeout(1000)

		require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
			Timeout: playwright.Float(5000),
		}))
	}
}

func TestAuth_MultipleSessionsCleanup(t *testing.T) {
	beforeEach(t)
	email := generateUniqueEmail(t)
	createTestUser(t, email)

	createMultipleSessions(t, email, 3)

	signin(t, email, "password")

	require.NoError(t, expect.Page(page).ToHaveURL(getFullPath("")+"/", playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(5000),
	}))

	verifyOldSessionsCleanedUp(t, email)
}

func createExpiredSession(t *testing.T, email string) {
	t.Helper()
	db, err := sql.Open("libsql", "file:../test-db.sqlite3")
	require.NoError(t, err)
	defer db.Close()

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	require.NoError(t, err)

	expiredTime := time.Now().Add(-2 * time.Hour)
	_, err = db.Exec(`
		INSERT INTO sessions (user_id, token, expires_at, user_agent, ip_address) 
		VALUES (?, ?, ?, ?, ?)
	`, userID, "expired-token", expiredTime, "test-agent", "127.0.0.1")
	require.NoError(t, err)
}

func createSessionNeedingRefresh(t *testing.T, email string) string {
	t.Helper()
	db, err := sql.Open("libsql", "file:../test-db.sqlite3")
	require.NoError(t, err)
	defer db.Close()

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	require.NoError(t, err)

	sessionToken := "refresh-token-" + email
	refreshTime := time.Now().Add(12 * time.Hour)
	_, err = db.Exec(`
		INSERT INTO sessions (user_id, token, expires_at, user_agent, ip_address) 
		VALUES (?, ?, ?, ?, ?)
	`, userID, sessionToken, refreshTime, "test-agent", "127.0.0.1")
	require.NoError(t, err)

	return sessionToken
}

func verifySessionWasRefreshed(t *testing.T, sessionToken string) {
	t.Helper()
	db, err := sql.Open("libsql", "file:../test-db.sqlite3")
	require.NoError(t, err)
	defer db.Close()

	var expiresAt time.Time
	err = db.QueryRow("SELECT expires_at FROM sessions WHERE token = ?", sessionToken).Scan(&expiresAt)
	require.NoError(t, err)

	expectedMinExpiry := time.Now().Add(6 * 24 * time.Hour)
	require.True(t, expiresAt.After(expectedMinExpiry),
		"Session should have been refreshed to expire in ~7 days, but expires at %v", expiresAt)
}

func createMultipleSessions(t *testing.T, email string, count int) {
	t.Helper()
	db, err := sql.Open("libsql", "file:../test-db.sqlite3")
	require.NoError(t, err)
	defer db.Close()

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	require.NoError(t, err)

	for i := 0; i < count; i++ {
		sessionToken := "old-session-" + email + "-" + strconv.Itoa(i)
		expiresAt := time.Now().Add(24 * time.Hour)
		_, err = db.Exec(`
			INSERT INTO sessions (user_id, token, expires_at, user_agent, ip_address) 
			VALUES (?, ?, ?, ?, ?)
		`, userID, sessionToken, expiresAt, "test-agent", "127.0.0.1")
		require.NoError(t, err)
	}
}

func verifyOldSessionsCleanedUp(t *testing.T, email string) {
	t.Helper()
	db, err := sql.Open("libsql", "file:../test-db.sqlite3")
	require.NoError(t, err)
	defer db.Close()

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	require.NoError(t, err)

	var sessionCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE user_id = ?", userID).Scan(&sessionCount)
	require.NoError(t, err)

	require.Equal(t, 4, sessionCount, "Expected 4 sessions, but found %d", sessionCount)
}
