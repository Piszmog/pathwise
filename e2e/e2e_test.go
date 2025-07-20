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
	"os/signal"
	"path/filepath"
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
	binaryPath  string
	appPort     int
	dbPath      string
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
	// Set up signal handler for cleanup on interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nReceived interrupt signal, cleaning up...")
		afterAll()
		os.Exit(1)
	}()

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
	// init web-first assertions with 5s timeout for faster tests
	expect = playwright.NewPlaywrightAssertions(5000)
	isChromium = browserName == "chromium" || browserName == ""
	isFirefox = browserName == "firefox"
	isWebKit = browserName == "webkit"

	// build and start app
	if err = buildApp(); err != nil {
		log.Fatalf("could not build app: %v", err)
	}
	if err = startApp(); err != nil {
		log.Fatalf("could not start app: %v", err)
	}
	if err = waitForAppReady(); err != nil {
		log.Fatalf("app did not become ready: %v", err)
	}
	// Give the app a moment to run migrations
	time.Sleep(2 * time.Second)
	if err = seedDB(); err != nil {
		log.Fatalf("could not seed db: %v", err)
	}
}

func buildApp() error {
	// Get absolute path to avoid cleanup issues
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	binaryPath = filepath.Join(filepath.Dir(wd), "pathwise-test")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	buildCmd.Dir = "../"
	return buildCmd.Run()
}

func startApp() error {
	appPort = getPort()

	// Get absolute path for database
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	dbPath = filepath.Join(filepath.Dir(wd), fmt.Sprintf("test-db-%d.sqlite3", appPort))

	app = exec.Command(binaryPath)
	app.Dir = "../"
	app.Env = append(
		os.Environ(),
		fmt.Sprintf("DB_URL=%s", dbPath),
		fmt.Sprintf("PORT=%d", appPort),
		"LOG_LEVEL=ERROR",
	)

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
	fmt.Printf("Started app on port %d, pid %d", appPort, app.Process.Pid)

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

func waitForAppReady() error {
	baseUrL, _ = url.Parse(fmt.Sprintf("http://localhost:%d", appPort))

	for i := 0; i < 30; i++ {
		resp, err := http.Get(baseUrL.String() + "/signin")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("app did not become ready within 3 seconds")
}

func cleanDB() error {
	// Open the same database file that the app is using
	db, err := sql.Open("libsql", fmt.Sprintf("file:%s", dbPath))
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing data in reverse dependency order
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
		if _, err := tx.Exec(query); err != nil {
			// Ignore errors for tables that might not exist yet
			continue
		}
	}

	return tx.Commit()
}
func seedDB() error {
	// Open the same database file that the app is using
	db, err := sql.Open("libsql", fmt.Sprintf("file:%s", dbPath))
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
	if binaryPath != "" {
		if err := os.Remove(binaryPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("could not remove binary %s: %v\n", binaryPath, err)
		}
	}
	if dbPath != "" {
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("could not remove database %s: %v\n", dbPath, err)
		}
	}
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
