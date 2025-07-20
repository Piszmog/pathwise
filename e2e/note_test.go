//go:build e2e

package e2e_test

import (
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestAddNote(t *testing.T) {
	beforeEach(t)

	user := useBaseUser(t, 1)
	signin(t, user.Email, "password")

	// Add a job application first
	addJobApplication(t, "Test Company", "Software Engineer", "https://example.com")

	// Open job details to access note form
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job details to load
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Add a note
	noteText := "This is a test note for the job application"
	addNote(t, noteText)

	// Verify note appears in timeline
	require.NoError(t, expect.Locator(page.GetByText("Note added")).ToBeVisible())
	require.NoError(t, expect.Locator(page.GetByText(noteText)).ToBeVisible())

	// Verify note form is cleared after submission
	require.NoError(t, expect.Locator(page.GetByPlaceholder("Add a note...")).ToHaveValue(""))
}

func TestAddMultipleNotes(t *testing.T) {
	beforeEach(t)

	user := useBaseUser(t, 1)
	signin(t, user.Email, "password")

	// Add a job application first
	addJobApplication(t, "Multi Note Company", "Developer", "https://example.com")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job details to load
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Add first note
	firstNote := "First note about the interview process"
	addNote(t, firstNote)

	// Verify first note was added before proceeding
	require.NoError(t, expect.Locator(page.GetByText(firstNote)).ToBeVisible())

	// Add second note
	secondNote := "Second note about salary discussion"
	addNote(t, secondNote)

	// Verify second note was added before proceeding
	require.NoError(t, expect.Locator(page.GetByText(secondNote)).ToBeVisible())

	// Add third note
	thirdNote := "Third note about team culture"
	addNote(t, thirdNote)

	// Verify third note was added
	require.NoError(t, expect.Locator(page.GetByText(thirdNote)).ToBeVisible())

	// Verify all notes appear in timeline
	require.NoError(t, expect.Locator(page.GetByText(firstNote)).ToBeVisible())
	require.NoError(t, expect.Locator(page.GetByText(secondNote)).ToBeVisible())
	require.NoError(t, expect.Locator(page.GetByText(thirdNote)).ToBeVisible())

	// Verify we have 3 "Note added" entries
	require.NoError(t, expect.Locator(page.GetByText("Note added")).ToHaveCount(3))

	// Verify all notes are present in timeline (includes 1 initial status + 3 notes = 4 total)
	timelineItems := page.Locator("#timeline-list li")
	require.NoError(t, expect.Locator(timelineItems).ToHaveCount(4))
}

func TestNoteWithSpecialCharacters(t *testing.T) {
	beforeEach(t)

	user := useBaseUser(t, 1)
	signin(t, user.Email, "password")

	// Add a job application first
	addJobApplication(t, "Special Chars Company", "QA Engineer", "https://example.com")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job details to load
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Test note with special characters
	specialNote := "Note with special chars: @#$%^&*()_+-=[]{}|;':\",./<>? and émojis 🚀 💻 ✅"
	addNote(t, specialNote)

	// Verify special characters are preserved and displayed correctly
	require.NoError(t, expect.Locator(page.GetByText(specialNote)).ToBeVisible())
}

func TestLongNote(t *testing.T) {
	beforeEach(t)

	user := useBaseUser(t, 1)
	signin(t, user.Email, "password")

	// Add a job application first
	addJobApplication(t, "Long Note Company", "Senior Developer", "https://example.com")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job details to load
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Test with a very long note
	longNote := "This is a very long note that contains multiple sentences and should test how the application handles longer text content. " +
		"It includes details about the interview process, the company culture, the technical challenges discussed, " +
		"the team dynamics, the project requirements, the technology stack, the development methodology, " +
		"the career growth opportunities, the compensation package, the work-life balance, " +
		"and any other relevant information that might be important for tracking this job application. " +
		"This note should be properly stored and displayed without any truncation or formatting issues."

	addNote(t, longNote)

	// Verify long note is displayed correctly
	require.NoError(t, expect.Locator(page.GetByText(longNote)).ToBeVisible())
}

func TestNotePersistenceAcrossPageRefresh(t *testing.T) {
	beforeEach(t)

	user := useBaseUser(t, 1)
	signin(t, user.Email, "password")

	// Add a job application first
	addJobApplication(t, "Persistence Company", "Full Stack Developer", "https://example.com")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job details to load
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Add a note
	persistentNote := "This note should persist after page refresh"
	addNote(t, persistentNote)

	// Verify note is visible
	require.NoError(t, expect.Locator(page.GetByText(persistentNote)).ToBeVisible())

	// Refresh the page
	_, err := page.Reload()
	require.NoError(t, err)

	// Navigate back to job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job details to load
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Verify note is still there after refresh
	require.NoError(t, expect.Locator(page.GetByText(persistentNote)).ToBeVisible())
	require.NoError(t, expect.Locator(page.GetByText("Note added")).ToBeVisible())
}

func TestEmptyNoteValidation(t *testing.T) {
	beforeEach(t)

	user := useBaseUser(t, 1)
	signin(t, user.Email, "password")

	// Add a job application first
	addJobApplication(t, "Validation Company", "Backend Developer", "https://example.com")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job details to load
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Try to submit empty note
	require.NoError(t, page.Locator("#note-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"}).Click())

	// Verify that HTML5 validation prevents submission (note field is required)
	// The form should not submit and no new timeline entry should appear
	require.NoError(t, expect.Locator(page.GetByText("Note added")).ToHaveCount(0))
}

func TestNoteTimelineIntegrationWithStatusChanges(t *testing.T) {
	beforeEach(t)

	user := useBaseUser(t, 1)
	signin(t, user.Email, "password")

	// Add a job application first
	addJobApplication(t, "Timeline Company", "DevOps Engineer", "https://example.com")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job details to load
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Add initial note
	firstNote := "Initial application submitted"
	addNote(t, firstNote)

	// Change status to interviewing
	updateJobApplication(t, "", "", "", "interviewing")

	// Add another note
	secondNote := "Had first round interview"
	addNote(t, secondNote)

	// Change status to offered
	updateJobApplication(t, "", "", "", "offered")

	// Add final note
	thirdNote := "Received job offer"
	addNote(t, thirdNote)

	// Verify timeline shows mixed entries (notes and status changes)
	require.NoError(t, expect.Locator(page.GetByText("Note added")).ToHaveCount(3))
	require.NoError(t, expect.Locator(page.GetByText("Status Change")).ToHaveCount(3)) // initial applied + interviewing + offered

	// Verify all notes are visible
	require.NoError(t, expect.Locator(page.GetByText(firstNote)).ToBeVisible())
	require.NoError(t, expect.Locator(page.GetByText(secondNote)).ToBeVisible())
	require.NoError(t, expect.Locator(page.GetByText(thirdNote)).ToBeVisible())

	// Verify status badges are visible in timeline
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("Interviewing")).ToBeVisible())
	require.NoError(t, expect.Locator(page.Locator("#timeline-list").GetByText("Offered")).ToBeVisible())
}

func TestNoteFormDisabledForArchivedJobs(t *testing.T) {
	beforeEach(t)

	user := useBaseUser(t, 1)
	signin(t, user.Email, "password")

	// Add a job application first
	addJobApplication(t, "Archive Note Company", "Frontend Developer", "https://example.com")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job details to load
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Add a note before archiving
	preArchiveNote := "Note added before archiving"
	addNote(t, preArchiveNote)

	// Archive the job
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"}).Click())
	waitForHTMXRequest(t)

	// Navigate to archives
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Archives"}).Click())

	// Open archived job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())

	// Wait for job details to load
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Verify note form is disabled (check if form exists first)
	noteForm := page.GetByPlaceholder("Add a note...")
	if count, _ := noteForm.Count(); count > 0 {
		require.NoError(t, expect.Locator(noteForm).ToBeDisabled())
		require.NoError(t, expect.Locator(page.Locator("#note-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"})).ToBeDisabled())
	} else {
		// Note form might not be present for archived jobs, which is also valid
		require.NoError(t, expect.Locator(page.Locator("#note-form")).ToHaveCount(0))
	}

	// Verify existing note is still visible
	require.NoError(t, expect.Locator(page.GetByText(preArchiveNote)).ToBeVisible())
}

func TestNoteFormReenabledAfterUnarchive(t *testing.T) {
	beforeEach(t)

	user := useBaseUser(t, 1)
	signin(t, user.Email, "password")

	// Add a job application first
	addJobApplication(t, "Unarchive Note Company", "Mobile Developer", "https://example.com")

	// Open job details and archive
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Archive"}).Click())
	waitForHTMXRequest(t)

	// Navigate to archives and unarchive
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Archives"}).Click())
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))
	require.NoError(t, page.Locator("#job-details").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Unarchive"}).Click())
	waitForHTMXRequest(t)

	// After unarchiving, the job is removed from archives page
	// Navigate back to main jobs page to find the unarchived job
	_, err := page.Goto(getFullPath(""))
	require.NoError(t, err)

	// Open the unarchived job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Verify note form is enabled again
	require.NoError(t, expect.Locator(page.GetByPlaceholder("Add a note...")).ToBeEnabled())
	require.NoError(t, expect.Locator(page.Locator("#note-form").GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Add"})).ToBeEnabled())

	// Add a note to confirm functionality
	postUnarchiveNote := "Note added after unarchiving"
	addNote(t, postUnarchiveNote)

	// Verify note was added successfully
	require.NoError(t, expect.Locator(page.GetByText(postUnarchiveNote)).ToBeVisible())
}

func TestNoteTimestampDisplay(t *testing.T) {
	beforeEach(t)

	user := useBaseUser(t, 1)
	signin(t, user.Email, "password")

	// Add a job application first
	addJobApplication(t, "Timestamp Company", "Data Scientist", "https://example.com")

	// Open job details
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "View job"}).First().Click())
	require.NoError(t, expect.Locator(page.Locator("#job-details")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(5000),
	}))

	// Add a note
	timestampNote := "Note to test timestamp display"
	addNote(t, timestampNote)

	// Verify timestamp is displayed (should show today's date)
	today := time.Now().Format("Mon Jan 2 2006")
	require.NoError(t, expect.Locator(page.GetByText("Note added").Locator("..").GetByText(today)).ToBeVisible())
}
