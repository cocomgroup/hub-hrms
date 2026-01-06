package e2e

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite is the end-to-end test suite
type E2ETestSuite struct {
	suite.Suite
	pw      *playwright.Playwright
	browser playwright.Browser
	baseURL string
}

// SetupSuite runs once before all tests
func (suite *E2ETestSuite) SetupSuite() {
	var err error
	suite.pw, err = playwright.Run()
	assert.NoError(suite.T(), err)

	suite.browser, err = suite.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	assert.NoError(suite.T(), err)

	// Set base URL from environment or default
	suite.baseURL = getEnv("E2E_BASE_URL", "http://localhost:5173")
}

// TearDownSuite runs once after all tests
func (suite *E2ETestSuite) TearDownSuite() {
	if suite.browser != nil {
		suite.browser.Close()
	}
	if suite.pw != nil {
		suite.pw.Stop()
	}
}

// createPage creates a new browser page
func (suite *E2ETestSuite) createPage() playwright.Page {
	page, err := suite.browser.NewPage()
	assert.NoError(suite.T(), err)
	return page
}

// TestUserRegistration tests the user registration flow
func (suite *E2ETestSuite) TestUserRegistration() {
	page := suite.createPage()
	defer page.Close()

	// Navigate to registration page
	_, err := page.Goto(suite.baseURL + "/register")
	assert.NoError(suite.T(), err)

	// Fill registration form
	timestamp := time.Now().Unix()
	email := fmt.Sprintf("test%d@example.com", timestamp)

	err = page.Fill("#email", email)
	assert.NoError(suite.T(), err)

	err = page.Fill("#first_name", "Test")
	assert.NoError(suite.T(), err)

	err = page.Fill("#last_name", "User")
	assert.NoError(suite.T(), err)

	err = page.Fill("#password", "TestPass123!")
	assert.NoError(suite.T(), err)

	err = page.Fill("#confirm_password", "TestPass123!")
	assert.NoError(suite.T(), err)

	// Submit form
	err = page.Click("button[type='submit']")
	assert.NoError(suite.T(), err)

	// Wait for redirect to dashboard
	err = page.WaitForURL(suite.baseURL+"/dashboard", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(5000),
	})
	assert.NoError(suite.T(), err)

	// Verify dashboard loaded
	content, err := page.TextContent("h1")
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), content, "Dashboard")
}

// TestUserLogin tests the login flow
func (suite *E2ETestSuite) TestUserLogin() {
	page := suite.createPage()
	defer page.Close()

	// Create test user first
	suite.createTestUser(page)

	// Navigate to login page
	_, err := page.Goto(suite.baseURL + "/login")
	assert.NoError(suite.T(), err)

	// Fill login form
	err = page.Fill("#email", "test@example.com")
	assert.NoError(suite.T(), err)

	err = page.Fill("#password", "TestPass123!")
	assert.NoError(suite.T(), err)

	// Submit form
	err = page.Click("button[type='submit']")
	assert.NoError(suite.T(), err)

	// Wait for redirect to dashboard
	err = page.WaitForURL(suite.baseURL+"/dashboard", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(5000),
	})
	assert.NoError(suite.T(), err)

	// Verify user is logged in
	name, err := page.TextContent(".user-name")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test User", name)
}

// TestInvalidLogin tests login with invalid credentials
func (suite *E2ETestSuite) TestInvalidLogin() {
	page := suite.createPage()
	defer page.Close()

	_, err := page.Goto(suite.baseURL + "/login")
	assert.NoError(suite.T(), err)

	err = page.Fill("#email", "nonexistent@example.com")
	assert.NoError(suite.T(), err)

	err = page.Fill("#password", "wrongpassword")
	assert.NoError(suite.T(), err)

	err = page.Click("button[type='submit']")
	assert.NoError(suite.T(), err)

	// Wait for error message
	err = page.WaitForSelector(".error-message", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(3000),
	})
	assert.NoError(suite.T(), err)

	errorText, err := page.TextContent(".error-message")
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorText, "Invalid credentials")
}

// TestTimesheetEntry tests creating a timesheet entry
func (suite *E2ETestSuite) TestTimesheetEntry() {
	page := suite.createPage()
	defer page.Close()

	// Login first
	suite.loginTestUser(page)

	// Navigate to timesheet page
	_, err := page.Goto(suite.baseURL + "/timesheets")
	assert.NoError(suite.T(), err)

	// Click add entry button
	err = page.Click("button#add-entry")
	assert.NoError(suite.T(), err)

	// Wait for modal
	err = page.WaitForSelector("#entry-modal", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	assert.NoError(suite.T(), err)

	// Fill entry form
	err = page.Fill("#date", "2024-01-15")
	assert.NoError(suite.T(), err)

	err = page.Fill("#hours", "8")
	assert.NoError(suite.T(), err)

	err = page.Fill("#description", "Development work")
	assert.NoError(suite.T(), err)

	err = page.SelectOption("#project", playwright.SelectOptionValues{
		Values: &[]string{"proj-1"},
	})
	assert.NoError(suite.T(), err)

	// Submit form
	err = page.Click("button#save-entry")
	assert.NoError(suite.T(), err)

	// Wait for success message
	err = page.WaitForSelector(".success-message", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(3000),
	})
	assert.NoError(suite.T(), err)

	// Verify entry appears in list
	entryText, err := page.TextContent(".timesheet-entry:last-child")
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), entryText, "Development work")
	assert.Contains(suite.T(), entryText, "8")
}

// TestManagerDashboard tests manager-specific features
func (suite *E2ETestSuite) TestManagerDashboard() {
	page := suite.createPage()
	defer page.Close()

	// Login as manager
	suite.loginAsManager(page)

	// Navigate to manager dashboard
	_, err := page.Goto(suite.baseURL + "/manager/dashboard")
	assert.NoError(suite.T(), err)

	// Verify manager-specific elements
	err = page.WaitForSelector("#team-overview")
	assert.NoError(suite.T(), err)

	err = page.WaitForSelector("#pending-approvals")
	assert.NoError(suite.T(), err)

	// Check team member count
	count, err := page.Locator(".team-member").Count()
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), count, 0)
}

// TestPaymentProcessing tests payment workflow
func (suite *E2ETestSuite) TestPaymentProcessing() {
	page := suite.createPage()
	defer page.Close()

	// Login as admin
	suite.loginAsAdmin(page)

	// Navigate to payments page
	_, err := page.Goto(suite.baseURL + "/payments")
	assert.NoError(suite.T(), err)

	// Select pay period
	err = page.SelectOption("#pay-period", playwright.SelectOptionValues{
		Values: &[]string{"2024-01"},
	})
	assert.NoError(suite.T(), err)

	// Wait for calculations to load
	err = page.WaitForSelector("#payment-summary", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	assert.NoError(suite.T(), err)

	// Verify total amount
	totalText, err := page.TextContent("#total-amount")
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), totalText)

	// Process payment
	err = page.Click("button#process-payment")
	assert.NoError(suite.T(), err)

	// Wait for confirmation modal
	err = page.WaitForSelector("#confirmation-modal", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	assert.NoError(suite.T(), err)

	// Confirm payment
	err = page.Click("button#confirm-payment")
	assert.NoError(suite.T(), err)

	// Wait for success
	err = page.WaitForSelector(".success-message", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(10000),
	})
	assert.NoError(suite.T(), err)
}

// TestResponsiveLayout tests responsive design
func (suite *E2ETestSuite) TestResponsiveLayout() {
	page := suite.createPage()
	defer page.Close()

	suite.loginTestUser(page)

	// Test mobile view
	err := page.SetViewportSize(375, 667)
	assert.NoError(suite.T(), err)

	_, err = page.Goto(suite.baseURL + "/dashboard")
	assert.NoError(suite.T(), err)

	// Check mobile menu
	err = page.WaitForSelector(".mobile-menu-button")
	assert.NoError(suite.T(), err)

	err = page.Click(".mobile-menu-button")
	assert.NoError(suite.T(), err)

	// Verify menu opens
	err = page.WaitForSelector(".mobile-nav", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	assert.NoError(suite.T(), err)

	// Test tablet view
	err = page.SetViewportSize(768, 1024)
	assert.NoError(suite.T(), err)

	err = page.Reload()
	assert.NoError(suite.T(), err)

	// Test desktop view
	err = page.SetViewportSize(1920, 1080)
	assert.NoError(suite.T(), err)

	err = page.Reload()
	assert.NoError(suite.T(), err)

	// Desktop menu should be visible
	err = page.WaitForSelector(".desktop-nav", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	assert.NoError(suite.T(), err)
}

// Helper functions

func (suite *E2ETestSuite) createTestUser(page playwright.Page) {
	// Implementation to create a test user
}

func (suite *E2ETestSuite) loginTestUser(page playwright.Page) {
	page.Goto(suite.baseURL + "/login")
	page.Fill("#email", "test@example.com")
	page.Fill("#password", "TestPass123!")
	page.Click("button[type='submit']")
	page.WaitForURL(suite.baseURL + "/dashboard")
}

func (suite *E2ETestSuite) loginAsManager(page playwright.Page) {
	page.Goto(suite.baseURL + "/login")
	page.Fill("#email", "manager@example.com")
	page.Fill("#password", "ManagerPass123!")
	page.Click("button[type='submit']")
	page.WaitForURL(suite.baseURL + "/dashboard")
}

func (suite *E2ETestSuite) loginAsAdmin(page playwright.Page) {
	page.Goto(suite.baseURL + "/login")
	page.Fill("#email", "admin@example.com")
	page.Fill("#password", "AdminPass123!")
	page.Click("button[type='submit']")
	page.WaitForURL(suite.baseURL + "/dashboard")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Run the test suite
func TestE2ETestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests")
	}
	suite.Run(t, new(E2ETestSuite))
}
