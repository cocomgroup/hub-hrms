package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// APIIntegrationTestSuite is the integration test suite
type APIIntegrationTestSuite struct {
	suite.Suite
	server    *httptest.Server
	client    *http.Client
	db        *sql.DB
	authToken string
}

// SetupSuite runs once before all tests
func (suite *APIIntegrationTestSuite) SetupSuite() {
	// Set test environment
	os.Setenv("ENV", "test")
	os.Setenv("DB_PATH", ":memory:")
	
	// Initialize test database
	db, err := InitTestDB()
	assert.NoError(suite.T(), err)
	suite.db = db
	
	// Run migrations
	err = RunMigrations(db)
	assert.NoError(suite.T(), err)
	
	// Create test server
	router := SetupRouter(db)
	suite.server = httptest.NewServer(router)
	suite.client = &http.Client{}
	
	// Create test user and get auth token
	suite.authToken = suite.createTestUserAndLogin()
}

// TearDownSuite runs once after all tests
func (suite *APIIntegrationTestSuite) TearDownSuite() {
	suite.server.Close()
	suite.db.Close()
}

// SetupTest runs before each test
func (suite *APIIntegrationTestSuite) SetupTest() {
	// Clean database tables before each test
	CleanDatabase(suite.db)
}

// createTestUserAndLogin creates a test user and returns auth token
func (suite *APIIntegrationTestSuite) createTestUserAndLogin() string {
	// Create user
	user := map[string]string{
		"email":      "test@example.com",
		"first_name": "Test",
		"last_name":  "User",
		"password":   "TestPass123!",
	}
	
	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", suite.server.URL+"/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	resp.Body.Close()
	
	// Login
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "TestPass123!",
	}
	
	body, _ = json.Marshal(loginData)
	req, _ = http.NewRequest("POST", suite.server.URL+"/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var loginResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	resp.Body.Close()
	
	return loginResp.Token
}

// TestUserRegistrationFlow tests the complete user registration flow
func (suite *APIIntegrationTestSuite) TestUserRegistrationFlow() {
	newUser := map[string]string{
		"email":      "newuser@example.com",
		"first_name": "New",
		"last_name":  "User",
		"password":   "NewPass123!",
	}
	
	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", suite.server.URL+"/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	
	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newUser["email"], user.Email)
	assert.NotEmpty(suite.T(), user.ID)
	resp.Body.Close()
}

// TestAuthenticationFlow tests login and token validation
func (suite *APIIntegrationTestSuite) TestAuthenticationFlow() {
	// Test successful login
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "TestPass123!",
	}
	
	body, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", suite.server.URL+"/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var loginResp struct {
		Token string `json:"token"`
		User  User   `json:"user"`
	}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), loginResp.Token)
	assert.Equal(suite.T(), "test@example.com", loginResp.User.Email)
	resp.Body.Close()
	
	// Test failed login with wrong password
	loginData["password"] = "WrongPassword"
	body, _ = json.Marshal(loginData)
	req, _ = http.NewRequest("POST", suite.server.URL+"/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

// TestUserCRUDOperations tests complete CRUD operations
func (suite *APIIntegrationTestSuite) TestUserCRUDOperations() {
	// Create
	newUser := map[string]string{
		"email":      "crud@example.com",
		"first_name": "CRUD",
		"last_name":  "Test",
		"password":   "CrudTest123!",
		"role":       "employee",
	}
	
	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", suite.server.URL+"/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	
	var createdUser User
	json.NewDecoder(resp.Body).Decode(&createdUser)
	resp.Body.Close()
	userID := createdUser.ID
	
	// Read
	req, _ = http.NewRequest("GET", suite.server.URL+"/api/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	resp, err = suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var retrievedUser User
	json.NewDecoder(resp.Body).Decode(&retrievedUser)
	assert.Equal(suite.T(), createdUser.Email, retrievedUser.Email)
	resp.Body.Close()
	
	// Update
	updateData := map[string]string{
		"first_name": "Updated",
		"last_name":  "Name",
	}
	
	body, _ = json.Marshal(updateData)
	req, _ = http.NewRequest("PUT", suite.server.URL+"/api/users/"+userID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	resp, err = suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// Verify update
	req, _ = http.NewRequest("GET", suite.server.URL+"/api/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	resp, err = suite.client.Do(req)
	json.NewDecoder(resp.Body).Decode(&retrievedUser)
	assert.Equal(suite.T(), "Updated", retrievedUser.FirstName)
	resp.Body.Close()
	
	// Delete
	req, _ = http.NewRequest("DELETE", suite.server.URL+"/api/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	resp, err = suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
	
	// Verify deletion
	req, _ = http.NewRequest("GET", suite.server.URL+"/api/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	resp, err = suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

// TestTimesheetOperations tests timesheet CRUD
func (suite *APIIntegrationTestSuite) TestTimesheetOperations() {
	// Create timesheet entry
	entry := map[string]interface{}{
		"date":        "2024-01-15",
		"hours":       8.0,
		"description": "Development work",
		"project_id":  "proj-123",
	}
	
	body, _ := json.Marshal(entry)
	req, _ := http.NewRequest("POST", suite.server.URL+"/api/timesheets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	
	var createdEntry TimesheetEntry
	json.NewDecoder(resp.Body).Decode(&createdEntry)
	assert.NotEmpty(suite.T(), createdEntry.ID)
	assert.Equal(suite.T(), 8.0, createdEntry.Hours)
	resp.Body.Close()
	
	// List timesheets
	req, _ = http.NewRequest("GET", suite.server.URL+"/api/timesheets?start_date=2024-01-01&end_date=2024-01-31", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	resp, err = suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var entries []TimesheetEntry
	json.NewDecoder(resp.Body).Decode(&entries)
	assert.NotEmpty(suite.T(), entries)
	resp.Body.Close()
}

// TestAuthorizationMiddleware tests authorization rules
func (suite *APIIntegrationTestSuite) TestAuthorizationMiddleware() {
	// Test without token
	req, _ := http.NewRequest("GET", suite.server.URL+"/api/users", nil)
	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
	
	// Test with invalid token
	req, _ = http.NewRequest("GET", suite.server.URL+"/api/users", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err = suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
	
	// Test with valid token
	req, _ = http.NewRequest("GET", suite.server.URL+"/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	resp, err = suite.client.Do(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

// Run the test suite
func TestAPIIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests")
	}
	suite.Run(t, new(APIIntegrationTestSuite))
}
