package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"banka1.com/dto"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

// MockRoundTripper implements http.RoundTripper for testing
type MockRoundTripper struct {
    Response *http.Response
    Error    error
}

func (m *MockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
    return m.Response, m.Error
}

// setupMockResponse creates a mock HTTP response
func setupMockResponse(statusCode int, body interface{}) *http.Response {
    return &http.Response{
        StatusCode: statusCode,
        Body:       httptest.NewRecorder().Result().Body, // Will be replaced in tests
        Header:     make(http.Header),
    }
}

// createFiberCtxWithToken creates a Fiber context with a token for testing
func createFiberCtxWithToken(token string) *fiber.Ctx {
    app := fiber.New()
    ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
    if token != "" {
        ctx.Locals("token", token)
    }
    return ctx
}

func TestGetEmployeesFiltered_Success(t *testing.T) {
    // Save original http.Client and restore it after test
    originalClient := http.DefaultClient
    defer func() { http.DefaultClient = originalClient }()

    // Set environment variable
    os.Setenv("USER_SERVICE", "http://mock-service")
    defer os.Unsetenv("USER_SERVICE")

    // Prepare mock response
    expectedResponse := dto.FilteredActuaryResponse{
        Data: []dto.FilteredActuaryDTO{
            {
                ID:             1,
                FirstName:      "John",
                LastName:       "Doe",
                Email:          "john.doe@example.com",
                Department:     "Finance",
                Position:       "Actuary",
                LimitAmount:    10000,
                UsedLimit:      5000,
            },
        },
    }

    // Setup test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request headers and query parameters
        assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
        assert.Equal(t, "John", r.URL.Query().Get("firstName"))
        assert.Equal(t, "Doe", r.URL.Query().Get("lastName"))
        assert.Equal(t, "john@example.com", r.URL.Query().Get("email"))
        assert.Equal(t, "Actuary", r.URL.Query().Get("position"))

        // Return successful response
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(expectedResponse)
    }))
    defer server.Close()

    os.Setenv("USER_SERVICE", server.URL)

    // Create test fiber context with token
    ctx := createFiberCtxWithToken("test-token")

    // Call the function we're testing
    employees, err := GetEmployeesFiltered(ctx, "John", "Doe", "john@example.com", "Actuary")

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, employees)
    assert.Equal(t, 1, len(employees))
    assert.Equal(t, uint(1), employees[0].ID)
    assert.Equal(t, "John", employees[0].FirstName)
    assert.Equal(t, "Doe", employees[0].LastName)
    assert.Equal(t, "john.doe@example.com", employees[0].Email)
}

func TestGetEmployeesFiltered_MissingToken(t *testing.T) {
    // Create context without token
    ctx := createFiberCtxWithToken("")

    // Call function
    employees, err := GetEmployeesFiltered(ctx, "", "", "", "")

    // Assertions
    assert.Error(t, err)
    assert.Nil(t, employees)
    assert.Contains(t, err.Error(), "token nije pronađen")
}

func TestGetEmployeesFiltered_MissingEnvironmentVariable(t *testing.T) {
    // Ensure environment variable is unset
    os.Unsetenv("USER_SERVICE")

    // Create context with token
    ctx := createFiberCtxWithToken("test-token")

    // Call function
    employees, err := GetEmployeesFiltered(ctx, "", "", "", "")

    // Assertions
    assert.Error(t, err)
    assert.Nil(t, employees)
    assert.Contains(t, err.Error(), "USER_SERVICE environment variable is not set")
}

func TestGetEmployeesFiltered_HttpError(t *testing.T) {
    // Set environment variable
    os.Setenv("USER_SERVICE", "http://invalid-service")
    defer os.Unsetenv("USER_SERVICE")

    // Create context with token
    ctx := createFiberCtxWithToken("test-token")

    // Call function
    employees, err := GetEmployeesFiltered(ctx, "", "", "", "")

    // Assertions
    assert.Error(t, err)
    assert.Nil(t, employees)
}

func TestGetEmployeesFiltered_NonOkStatus(t *testing.T) {
    // Set environment variable
    os.Setenv("USER_SERVICE", "http://mock-service")
    defer os.Unsetenv("USER_SERVICE")

    // Setup test server returning non-OK status
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusUnauthorized)
        w.Write([]byte("Unauthorized"))
    }))
    defer server.Close()

    os.Setenv("USER_SERVICE", server.URL)

    // Create context with token
    ctx := createFiberCtxWithToken("test-token")

    // Call function
    employees, err := GetEmployeesFiltered(ctx, "", "", "", "")

    // Assertions
    assert.Error(t, err)
    assert.Nil(t, employees)
    assert.Contains(t, err.Error(), "status: 401")
}

func TestGetEmployeesFiltered_InvalidJson(t *testing.T) {
    // Set environment variable
    os.Setenv("USER_SERVICE", "http://mock-service")
    defer os.Unsetenv("USER_SERVICE")

    // Setup test server returning invalid JSON
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("{invalid json}"))
    }))
    defer server.Close()

    os.Setenv("USER_SERVICE", server.URL)

    // Create context with token
    ctx := createFiberCtxWithToken("test-token")

    // Call function
    employees, err := GetEmployeesFiltered(ctx, "", "", "", "")

    // Assertions
    assert.Error(t, err)
    assert.Nil(t, employees)
}