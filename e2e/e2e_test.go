//go:build e2e

package e2e_test

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

// global variables, can be used in any tests
var (
	pw          *playwright.Playwright
	browser     playwright.Browser
	context     playwright.BrowserContext
	page        playwright.Page
	expect      playwright.PlaywrightAssertions
	isChromium  bool
	isFirefox   bool
	isWebKit    bool
	browserName = getBrowserName()
	browserType playwright.BrowserType
	app         *exec.Cmd
	baseUrL     *url.URL
)

// defaultContextOptions for most tests
var defaultContextOptions = playwright.BrowserNewContextOptions{
	AcceptDownloads: playwright.Bool(true),
	HasTouch:        playwright.Bool(true),
}

func TestMain(m *testing.M) {
	beforeAll()
	code := m.Run()
	afterAll()
	os.Exit(code)
}

// beforeAll prepares the environment, including
//   - start Playwright driver
//   - launch browser depends on BROWSER env
//   - init web-first assertions, alias as `expect`
func beforeAll() {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("could not install Playwright: %v", err)
	}

	pw, err = playwright.Run()
	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	switch browserName {
	case "firefox":
		browserType = pw.Firefox
	case "webkit":
		browserType = pw.WebKit
	case "chromium":
		fallthrough
	default:
		browserType = pw.Chromium
	}
	// launch browser, headless or not depending on HEADFUL env
	browser, err = browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(os.Getenv("HEADFUL") == ""),
	})
	if err != nil {
		log.Fatalf("could not launch: %v", err)
	}
	// init web-first assertions with 10s timeout for more reliable tests
	expect = playwright.NewPlaywrightAssertions(10000)
	isChromium = browserName == "chromium" || browserName == ""
	isFirefox = browserName == "firefox"
	isWebKit = browserName == "webkit"

	// start app
	if err = startApp(); err != nil {
		log.Fatalf("could not start app: %v", err)
	}

	// wait for server to be ready
	if err = waitForServer(); err != nil {
		log.Fatalf("could not wait for server: %v", err)
	}
}

func startApp() error {
	port := getPort()
	app = exec.Command("go", "run", "main.go")
	app.Dir = "../"
	app.Env = append(
		os.Environ(),
		"DB_URL=./test-db.sqlite3",
		fmt.Sprintf("PORT=%d", port),
		"LOG_LEVEL=DEBUG",
	)

	var err error
	baseUrL, err = url.Parse(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		return err
	}

	stdout, err := app.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := app.StderrPipe()
	if err != nil {
		return err
	}

	if err := app.Start(); err != nil {
		return err
	}
	fmt.Printf("Started app on port %d, pid %d", port, app.Process.Pid)

	stdoutchan := make(chan string)
	stderrchan := make(chan string)
	go func() {
		defer close(stdoutchan)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			stdoutchan <- scanner.Text()
		}
	}()
	go func() {
		defer close(stderrchan)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			stderrchan <- scanner.Text()
		}
	}()

	go func() {
		for line := range stdoutchan {
			fmt.Println("[STDOUT]", line)
		}
	}()
	go func() {
		for line := range stderrchan {
			fmt.Println("[STDERR]", line)
		}
	}()
	return nil
}

func waitForServer() error {
	for i := 0; i < 30; i++ {
		resp, err := http.Get(baseUrL.String() + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("server not ready after 30 seconds")
}

func cleanDB() error {
	db, err := sql.Open("libsql", "file:../test-db.sqlite3")
	if err != nil {
		return err
	}
	defer db.Close()

	// Clear existing data
	clearQueries := []string{
		"DELETE FROM job_application_notes;",
		"DELETE FROM job_application_status_histories;",
		"DELETE FROM job_applications;",
		"DELETE FROM job_application_stats;",
		"DELETE FROM sessions;",
		"DELETE FROM user_ips;",
		"DELETE FROM users;",
	}

	for _, query := range clearQueries {
		if _, err := db.Exec(query); err != nil {
			// Ignore errors for tables that might not exist yet
			continue
		}
	}
	return nil
}

func seedDB() error {
	db, err := sql.Open("libsql", "file:../test-db.sqlite3")
	if err != nil {
		return err
	}
	defer db.Close()

	b, err := os.ReadFile("./testdata/seed.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(b))
	if err != nil {
		return err
	}
	return nil
}
func getPort() int {
	randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
	return randomGenerator.Intn(9001-3000) + 3000
}

// afterAll does cleanup, e.g. stop playwright driver
func afterAll() {
	if app != nil && app.Process != nil {
		if err := syscall.Kill(-app.Process.Pid, syscall.SIGKILL); err != nil {
			fmt.Println(err)
		}
	}
	if err := pw.Stop(); err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	if err := removeDBFile(); err != nil {
		log.Fatalf("could not remove test-db.sqlite3: %v", err)
	}
}

func removeDBFile() error {
	return os.Remove("../test-db.sqlite3")
}

// beforeEach creates a new context and page for each test,
// so each test has isolated environment. Usage:
//
//	Func TestFoo(t *testing.T) {
//	  beforeEach(t)
//	  // your test code
//	}
func beforeEach(t *testing.T, contextOptions ...playwright.BrowserNewContextOptions) {
	t.Helper()
	opt := defaultContextOptions
	if len(contextOptions) == 1 {
		opt = contextOptions[0]
	}
	context, page = newBrowserContextAndPage(t, opt)

	// Wait for server to be ready
	if err := waitForServer(); err != nil {
		t.Fatalf("could not wait for server: %v", err)
	}

	// Clean database before each test to ensure isolation
	if err := cleanDB(); err != nil {
		t.Fatalf("could not clean db: %v", err)
	}
	if err := seedDB(); err != nil {
		t.Fatalf("could not seed db: %v", err)
	}
}

func getBrowserName() string {
	browserName, hasEnv := os.LookupEnv("BROWSER")
	if hasEnv {
		return browserName
	}
	return "chromium"
}

func newBrowserContextAndPage(t *testing.T, options playwright.BrowserNewContextOptions) (playwright.BrowserContext, playwright.Page) {
	t.Helper()
	context, err := browser.NewContext(options)
	if err != nil {
		t.Fatalf("could not create new context: %v", err)
	}
	t.Cleanup(func() {
		if ctxErr := context.Close(); ctxErr != nil {
			t.Errorf("could not close context: %v", ctxErr)
		}
	})
	p, err := context.NewPage()
	if err != nil {
		t.Fatalf("could not create new page: %v", err)
	}
	return context, p
}

func getFullPath(relativePath string) string {
	return baseUrL.ResolveReference(&url.URL{Path: relativePath}).String()
}

func createTestUser(t *testing.T, email, password string) {
	t.Helper()
	db, err := sql.Open("libsql", "file:../test-db.sqlite3")
	if err != nil {
		t.Fatalf("could not open db: %v", err)
	}
	defer db.Close()

	hashedPassword := "$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm"

	_, err = db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, hashedPassword)
	if err != nil {
		t.Fatalf("could not create test user: %v", err)
	}

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		t.Fatalf("could not get user ID: %v", err)
	}

	_, err = db.Exec("INSERT INTO job_application_stats (user_id) VALUES (?)", userID)
	if err != nil {
		t.Fatalf("could not create job application stats: %v", err)
	}
}

func generateUniqueEmail(t *testing.T) string {
	t.Helper()
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("test-%d@example.com", timestamp)
}

func resetUserData(t *testing.T, email string) {
	t.Helper()
	db, err := sql.Open("libsql", "file:../test-db.sqlite3")
	if err != nil {
		t.Fatalf("could not open db: %v", err)
	}
	defer db.Close()

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		return
	}

	clearQueries := []string{
		"DELETE FROM job_application_notes WHERE job_application_id IN (SELECT id FROM job_applications WHERE user_id = ?)",
		"DELETE FROM job_application_status_histories WHERE job_application_id IN (SELECT id FROM job_applications WHERE user_id = ?)",
		"DELETE FROM job_applications WHERE user_id = ?",
		"UPDATE job_application_stats SET total_applications = 0, total_companies = 0, total_applied = 0, total_interviewing = 0, total_offered = 0, total_rejected = 0 WHERE user_id = ?",
		"DELETE FROM sessions WHERE user_id = ?",
		"DELETE FROM user_ips WHERE user_id = ?",
	}

	for _, query := range clearQueries {
		if _, err := db.Exec(query, userID); err != nil {
			continue
		}
	}
}

func createUserAndSignIn(t *testing.T) string {
	t.Helper()
	email := generateUniqueEmail(t)
	createTestUser(t, email, "password")
	signin(t, email, "password")
	return email
}
