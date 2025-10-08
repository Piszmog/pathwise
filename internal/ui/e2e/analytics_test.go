//go:build e2e

package e2e_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestAnalytics_SingleStatusOnly(t *testing.T) {
	beforeEach(t)
	email := createUserAndSignIn(t)
	userID := getUserIDByEmail(t, email)

	createJobApplicationWithStatus(t, userID, "Tech Corp", "Software Engineer", "applied")
	createJobApplicationWithStatus(t, userID, "Another Corp", "Backend Dev", "applied")

	_, err := page.Goto(getFullPath("analytics"))
	require.NoError(t, err)

	waitForSankeyRender(t)

	sankeyContainer := page.Locator("#sankey-viz")
	svgElement := sankeyContainer.Locator("svg")

	nodes := svgElement.Locator("rect")
	nodeCount, err := nodes.Count()
	require.NoError(t, err)

	if nodeCount > 0 {
		textLabels := svgElement.Locator("text")
		require.NoError(t, expect.Locator(textLabels.Filter(playwright.LocatorFilterOptions{HasText: "applied"})).ToBeVisible())
	}
}

func TestAnalytics_StatusTransitions(t *testing.T) {
	beforeEach(t)
	email := createUserAndSignIn(t)
	userID := getUserIDByEmail(t, email)

	createJobApplicationWithHistory(t, userID, "Flow Corp", "Engineer", []string{"applied", "interviewing", "offered"})
	createJobApplicationWithHistory(t, userID, "Second Corp", "Developer", []string{"applied", "interviewing"})

	_, err := page.Goto(getFullPath("analytics"))
	require.NoError(t, err)

	waitForSankeyRender(t)

	sankeyContainer := page.Locator("#sankey-viz")
	svgElement := sankeyContainer.Locator("svg")

	nodes := svgElement.Locator("rect")
	nodeCount, err := nodes.Count()
	require.NoError(t, err)
	require.Greater(t, nodeCount, 0)

	links := svgElement.Locator("path")
	linkCount, err := links.Count()
	require.NoError(t, err)
	require.Greater(t, linkCount, 0)

	textLabels := svgElement.Locator("text")
	require.NoError(t, expect.Locator(textLabels.Filter(playwright.LocatorFilterOptions{HasText: "applied"})).ToBeVisible())
	require.NoError(t, expect.Locator(textLabels.Filter(playwright.LocatorFilterOptions{HasText: "interviewing"})).ToBeVisible())
	require.NoError(t, expect.Locator(textLabels.Filter(playwright.LocatorFilterOptions{HasText: "offered"})).ToBeVisible())
}

func TestAnalytics_ComplexWorkflow(t *testing.T) {
	beforeEach(t)
	email := createUserAndSignIn(t)
	userID := getUserIDByEmail(t, email)

	createAnalyticsTestData(t, userID)

	_, err := page.Goto(getFullPath("analytics"))
	require.NoError(t, err)

	waitForSankeyRender(t)

	sankeyContainer := page.Locator("#sankey-viz")
	svgElement := sankeyContainer.Locator("svg")

	nodes := svgElement.Locator("rect")
	nodeCount, err := nodes.Count()
	require.NoError(t, err)
	require.Greater(t, nodeCount, 4)

	links := svgElement.Locator("path")
	linkCount, err := links.Count()
	require.NoError(t, err)
	require.Greater(t, linkCount, 2)

	textLabels := svgElement.Locator("text")
	expectedStatuses := []string{"applied", "interviewing", "rejected", "offered", "accepted"}
	for _, status := range expectedStatuses {
		require.NoError(t, expect.Locator(textLabels.Filter(playwright.LocatorFilterOptions{HasText: status})).ToBeVisible())
	}
}

func TestAnalytics_ResponsiveLayout(t *testing.T) {
	beforeEach(t)
	email := createUserAndSignIn(t)
	userID := getUserIDByEmail(t, email)

	createJobApplicationWithHistory(t, userID, "Responsive Corp", "Engineer", []string{"applied", "interviewing"})

	// Test desktop view
	page.SetViewportSize(1920, 1080)
	_, err := page.Goto(getFullPath("analytics"))
	require.NoError(t, err)

	waitForSankeyRender(t)

	sankeyContainer := page.Locator("#sankey-viz")
	require.NoError(t, expect.Locator(sankeyContainer).ToBeVisible())

	// Test mobile view
	page.SetViewportSize(375, 667)
	_, err = page.Reload()
	require.NoError(t, err)
	waitForSankeyRender(t)

	// Just verify it still renders on mobile
	require.NoError(t, expect.Locator(sankeyContainer).ToBeVisible())
}

func TestAnalytics_LoadingState(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	_, err := page.Goto(getFullPath("analytics"))
	require.NoError(t, err)

	// Check that the analytics page loads
	require.NoError(t, expect.Locator(page.Locator("#sankey-container")).ToBeVisible())

	// Check if loading text is present (it may load too fast to catch spinner)
	loadingText := page.GetByText("Loading analytics...")
	textCount, _ := loadingText.Count()
	if textCount > 0 {
		require.NoError(t, expect.Locator(loadingText).ToBeVisible())
	}

	waitForSankeyRender(t)

	// Verify final state - sankey viz should be visible
	require.NoError(t, expect.Locator(page.Locator("#sankey-viz")).ToBeVisible())
}

func createJobApplicationWithStatus(t *testing.T, userID int64, company, title, status string) {
	t.Helper()
	db, err := sql.Open("libsql", getDBURL(t))
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()

	var jobID int64
	err = tx.QueryRowContext(ctx,
		`INSERT INTO job_applications (company, title, url, applied_at, user_id, status, archived) 
		 VALUES (?, ?, ?, datetime('now'), ?, ?, 0) RETURNING id`,
		company, title, fmt.Sprintf("https://%s.com", company), userID, status).Scan(&jobID)
	require.NoError(t, err)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO job_application_status_histories (job_application_id, status, created_at) 
		 VALUES (?, ?, datetime('now'))`,
		jobID, status)
	require.NoError(t, err)

	require.NoError(t, tx.Commit())
}

func createJobApplicationWithHistory(t *testing.T, userID int64, company, title string, statusHistory []string) {
	t.Helper()
	db, err := sql.Open("libsql", getDBURL(t))
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()

	if len(statusHistory) == 0 {
		statusHistory = []string{"applied"}
	}

	var jobID int64
	err = tx.QueryRowContext(ctx,
		`INSERT INTO job_applications (company, title, url, applied_at, user_id, status, archived) 
		 VALUES (?, ?, ?, datetime('now'), ?, ?, 0) RETURNING id`,
		company, title, fmt.Sprintf("https://%s.com", company), userID, statusHistory[len(statusHistory)-1]).Scan(&jobID)
	require.NoError(t, err)

	for i, status := range statusHistory {
		createdAt := fmt.Sprintf("datetime('now', '-%d hours')", len(statusHistory)-i-1)
		_, err = tx.ExecContext(ctx,
			fmt.Sprintf(`INSERT INTO job_application_status_histories (job_application_id, status, created_at) 
			 VALUES (?, ?, %s)`, createdAt),
			jobID, status)
		require.NoError(t, err)
	}

	require.NoError(t, tx.Commit())
}

func createAnalyticsTestData(t *testing.T, userID int64) {
	t.Helper()

	createJobApplicationWithHistory(t, userID, "Success Corp", "Senior Engineer", []string{"applied", "interviewing", "offered", "accepted"})
	createJobApplicationWithHistory(t, userID, "Reject Corp", "Backend Dev", []string{"applied", "interviewing", "rejected"})
	createJobApplicationWithHistory(t, userID, "Decline Corp", "Frontend Dev", []string{"applied", "interviewing", "offered", "declined"})
	createJobApplicationWithHistory(t, userID, "Withdraw Corp", "DevOps Engineer", []string{"applied", "withdrawn"})
	createJobApplicationWithHistory(t, userID, "Pipeline Corp", "Full Stack", []string{"watching", "applied", "interviewing"})
	createJobApplicationWithStatus(t, userID, "Fresh Corp", "Junior Dev", "applied")
	createJobApplicationWithStatus(t, userID, "Wait Corp", "Designer", "watching")
}

func getUserIDByEmail(t *testing.T, email string) int64 {
	t.Helper()
	db, err := sql.Open("libsql", getDBURL(t))
	require.NoError(t, err)
	defer db.Close()

	var userID int64
	err = db.QueryRowContext(context.Background(), "SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	require.NoError(t, err)
	return userID
}

func waitForSankeyRender(t *testing.T) {
	t.Helper()

	sankeyContainer := page.Locator("#sankey-container")
	require.NoError(t, expect.Locator(sankeyContainer).ToBeVisible())

	sankeyViz := page.Locator("#sankey-viz")
	require.NoError(t, expect.Locator(sankeyViz).ToBeVisible())

	page.WaitForFunction("() => document.querySelector('#sankey-viz svg') !== null", playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(10000),
	})

	page.WaitForFunction("() => !document.querySelector('#sankey-container .animate-spin')", playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(10000),
	})

	time.Sleep(500 * time.Millisecond)
}

func TestAnalytics_ErrorHandling(t *testing.T) {
	beforeEach(t)
	createUserAndSignIn(t)

	_, err := page.Goto(getFullPath("analytics"))
	require.NoError(t, err)

	// For a signed-in user, the page should load successfully
	// Check that we get to the analytics page without errors
	require.NoError(t, expect.Locator(page.Locator("#sankey-container")).ToBeVisible())

	waitForSankeyRender(t)

	// Verify the visualization loads without errors
	sankeyViz := page.Locator("#sankey-viz")
	require.NoError(t, expect.Locator(sankeyViz).ToBeVisible())
}
