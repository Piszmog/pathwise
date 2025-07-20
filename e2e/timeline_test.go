//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestTimeline_MixedEntriesChronologicalOrder(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Timeline Company", "Software Engineer", "https://timeline.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Initial status should be "applied"
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("applied")).ToHaveCount(1))

	// Add a note
	addNote(t, "First note after application")
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("First note after application")).ToHaveCount(1))

	// Change status to interviewing
	updateJobApplication(t, "", "", "", "interviewing")
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("interviewing")).ToHaveCount(1))

	// Add another note
	addNote(t, "Had technical interview today")
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("Had technical interview today")).ToHaveCount(1))

	// Change status to offered
	updateJobApplication(t, "", "", "", "offered")
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("offered")).ToHaveCount(1))

	// Add final note
	addNote(t, "Received offer! Negotiating salary")
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("Received offer! Negotiating salary")).ToHaveCount(1))

	// Verify all entries are present (3 status changes + 3 notes = 6 total)
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(6))

	// Verify chronological order (newest first)
	timelineItems := page.Locator("#timeline-list > li")

	// First item should be the most recent note
	require.NoError(t, expect.Locator(timelineItems.Nth(0).GetByText("Received offer! Negotiating salary")).ToBeVisible())

	// Second item should be the offered status
	require.NoError(t, expect.Locator(timelineItems.Nth(1).GetByText("offered")).ToBeVisible())

	// Third item should be the technical interview note
	require.NoError(t, expect.Locator(timelineItems.Nth(2).GetByText("Had technical interview today")).ToBeVisible())

	// Fourth item should be the interviewing status
	require.NoError(t, expect.Locator(timelineItems.Nth(3).GetByText("interviewing")).ToBeVisible())

	// Fifth item should be the first note
	require.NoError(t, expect.Locator(timelineItems.Nth(4).GetByText("First note after application")).ToBeVisible())

	// Last item should be the initial applied status
	require.NoError(t, expect.Locator(timelineItems.Nth(5).GetByText("applied")).ToBeVisible())
}

func TestTimeline_StatusChangeCreatesEntry(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Status Timeline Company", "Backend Developer", "https://statustimeline.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Should start with initial "applied" status
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("applied")).ToHaveCount(1))

	// Each status change should add a new timeline entry
	statuses := []string{"watching", "interviewing", "offered", "accepted"}
	for i, status := range statuses {
		updateJobApplication(t, "", "", "", status)

		// Timeline should have initial + i+1 new entries
		expectedCount := i + 2
		require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(expectedCount))
		require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText(status)).ToHaveCount(1))
	}

	// Final count should be 5 (initial applied + 4 status changes)
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(5))
}

func TestTimeline_NoteCreatesEntry(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Note Timeline Company", "Frontend Developer", "https://notetimeline.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Should start with initial "applied" status
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(1))

	// Add multiple notes
	notes := []string{
		"Applied through company website",
		"Recruiter reached out for phone screen",
		"Completed phone screen, moving to technical",
		"Technical interview scheduled for next week",
		"Completed technical, waiting for feedback",
	}

	for i, note := range notes {
		addNote(t, note)

		// Timeline should have initial status + i+1 notes
		expectedCount := i + 2
		require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(expectedCount))
		require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText(note)).ToHaveCount(1))
	}

	// Final count should be 6 (initial applied + 5 notes)
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(6))
}

func TestTimeline_EmptyTimelineForNewJob(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Empty Timeline Company", "DevOps Engineer", "https://emptytimeline.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// New job should have exactly one timeline entry (initial applied status)
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("applied")).ToHaveCount(1))
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("Note added")).ToHaveCount(0))
}

func TestTimeline_TimelineVisibilityToggle(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Toggle Timeline Company", "Full Stack Developer", "https://toggletimeline.com")

	// Timeline should not be visible on main page
	require.NoError(t, expect.Locator(page.Locator("#timeline")).ToHaveCount(0))

	// Open job details to show timeline
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	require.NoError(t, expect.Locator(page.Locator("#timeline")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(1))

	// Close job details (navigate away)
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Pathwise"}).Click())
	require.NoError(t, expect.Locator(page.Locator("#timeline")).ToHaveCount(0))
}

func TestTimeline_LargeTimelinePerformance(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Large Timeline Company", "Performance Engineer", "https://largetimeline.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Add many notes to test performance
	for i := 1; i <= 20; i++ {
		addNote(t, "Note number "+string(rune(i+'0')))
	}

	// Add several status changes
	statuses := []string{"watching", "interviewing", "offered", "rejected"}
	for _, status := range statuses {
		updateJobApplication(t, "", "", "", status)
	}

	// Should have 1 initial status + 20 notes + 4 status changes = 25 total
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(25))

	// Timeline should still be responsive
	require.NoError(t, expect.Locator(page.Locator("#timeline")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("Note number 1")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("rejected")).ToBeVisible())
}

func TestTimeline_TimelineEntryTypes(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Entry Types Company", "QA Engineer", "https://entrytypes.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Add a note
	addNote(t, "This is a note entry")

	// Change status
	updateJobApplication(t, "", "", "", "interviewing")

	// Verify different entry types are distinguishable
	timelineItems := page.Locator("#timeline-list > li")
	require.NoError(t, expect.Locator(timelineItems).ToHaveCount(3)) // applied + note + interviewing

	// Check that status entries and note entries are visually different
	// Status entries should contain status text
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("interviewing")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("applied")).ToBeVisible())

	// Note entries should contain note text
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("This is a note entry")).ToBeVisible())
}

func TestTimeline_TimelineWithSpecialCharacters(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Special Chars Company", "Unicode Engineer", "https://specialchars.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Add notes with special characters
	specialNotes := []string{
		"Note with émojis 🚀 and ñ characters",
		"Note with <script>alert('xss')</script> HTML",
		"Note with \"quotes\" and 'apostrophes'",
		"Note with line\nbreaks and\ttabs",
		"Note with unicode: ∑∆∏∫∂∇",
	}

	for _, note := range specialNotes {
		addNote(t, note)
		require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText(note)).ToBeVisible())
	}

	// Should have initial status + 5 special notes = 6 total
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(6))
}

func TestTimeline_TimelinePersistenceAcrossPageRefresh(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Persistence Timeline Company", "Data Engineer", "https://persistencetimeline.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Add some timeline entries
	addNote(t, "First note before refresh")
	updateJobApplication(t, "", "", "", "interviewing")
	addNote(t, "Second note before refresh")

	// Should have 4 entries: applied + note + interviewing + note
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(4))

	// Refresh the page
	_, err := page.Reload()
	require.NoError(t, err)

	// Re-signin after refresh
	signin(t, user.Email, "password")

	// Open job details again
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Timeline should persist
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(4))
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("First note before refresh")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("Second note before refresh")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("interviewing")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("applied")).ToBeVisible())
}

func TestTimeline_ArchivedJobTimelineReadOnly(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Archive Timeline Company", "Security Engineer", "https://archivetimeline.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Add some timeline entries before archiving
	addNote(t, "Note before archiving")
	updateJobApplication(t, "", "", "", "interviewing")

	// Archive the job
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"}).Click())
	waitForHTMXRequest(t)

	// Navigate to archives
	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)

	// Open archived job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Timeline should still be visible and contain all entries
	require.NoError(t, expect.Locator(page.Locator("#timeline")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(3)) // applied + note + interviewing
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("Note before archiving")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("interviewing")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("applied")).ToBeVisible())

	// Note form should be disabled for archived jobs
	require.NoError(t, expect.Locator(page.GetByPlaceholder("Add a note...")).ToBeDisabled())
}

func TestTimeline_UnarchiveJobTimelineEditable(t *testing.T) {
	beforeEach(t)
	user := createTestUser(t, "timeline")
	signin(t, user.Email, "password")

	addJobApplication(t, "Unarchive Timeline Company", "Platform Engineer", "https://unarchivetimeline.com")
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Add timeline entries and archive
	addNote(t, "Note before archiving")
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"}).Click())
	waitForHTMXRequest(t)

	// Navigate to archives and unarchive
	_, err := page.Goto(getFullPath("archives"))
	require.NoError(t, err)
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Unarchive"}).Click())
	waitForHTMXRequest(t)

	// Navigate back to main page
	_, err = page.Goto(getFullPath(""))
	require.NoError(t, err)
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Timeline should be editable again
	require.NoError(t, expect.Locator(page.GetByPlaceholder("Add a note...")).ToBeEnabled())

	// Should be able to add new timeline entries
	addNote(t, "Note after unarchiving")
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("Note after unarchiving")).ToBeVisible())

	// Should have 3 entries: applied + note before + note after
	require.NoError(t, expect.Locator(page.Locator("#timeline-list > li")).ToHaveCount(3))
}
