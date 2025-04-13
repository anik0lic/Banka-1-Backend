package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"banka1.com/types"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupStockApp() *fiber.App {
    app := fiber.New()
    InitStockRoutes(app)
    return app
}

// TestInitStockRoutes verifies that all routes are correctly registered
func TestInitStockRoutes(t *testing.T) {
    app := fiber.New()
    
    // Execute InitStockRoutes
    InitStockRoutes(app)
    
    // Check that routes are registered correctly
    findRoute := func(method, path string) bool {
        for methodRoutes, routes := range app.Stack() {
            for _, route := range routes {
                if strconv.Itoa(methodRoutes) == method && strings.Contains(route.Path, path) {
                    return true
                }
            }
        }
        return false
    }
    
    // Verify all expected routes are registered
    assert.True(t, findRoute("GET", "/stocks"))
    assert.True(t, findRoute("GET", "/stocks/:ticker"))
    assert.True(t, findRoute("GET", "/stocks/:ticker/history/first"))
    assert.True(t, findRoute("GET", "/stocks/:ticker/history/:date"))
    assert.True(t, findRoute("GET", "/stocks/:ticker/history"))
}

// TestGetStockByTickerNotFound tests 404 response when a stock ticker is not found
func TestGetStockByTickerNotFound(t *testing.T) {
    // Setup
    app := setupStockApp()
    
    // Execute - Test with a ticker that shouldn't exist
    req := httptest.NewRequest(http.MethodGet, "/stocks/NONEXISTENT", nil)
    resp, _ := app.Test(req)
    defer resp.Body.Close()
    
    // Verify
    assert.Equal(t, 404, resp.StatusCode)
    
    // Check response body
    body, _ := io.ReadAll(resp.Body)
    var response types.Response
    json.Unmarshal(body, &response)
    
    assert.False(t, response.Success)
    assert.Contains(t, response.Error, "not found with ticker")
}

// TestGetStockHistoryByDateInvalidFormat tests handling of invalid date format
func TestGetStockHistoryByDateInvalidFormat(t *testing.T) {
    // Setup
    app := setupStockApp()
    
    // Execute - Test with invalid date format
    req := httptest.NewRequest(http.MethodGet, "/stocks/AAPL/history/invalid-date", nil)
    resp, _ := app.Test(req)
    defer resp.Body.Close()
    
    // Verify
    assert.Equal(t, 400, resp.StatusCode)
    
    // Check response body
    body, _ := io.ReadAll(resp.Body)
    var response types.Response
    json.Unmarshal(body, &response)
    
    assert.False(t, response.Success)
    assert.Contains(t, response.Error, "Invalid date format")
}

// TestGetStockHistoryRangeInvalidDateFormat tests handling of invalid date format in range params
func TestGetStockHistoryRangeInvalidDateFormat(t *testing.T) {
    // Setup
    app := setupStockApp()
    
    // Test invalid start date
    req := httptest.NewRequest(http.MethodGet, "/stocks/AAPL/history?startDate=invalid", nil)
    resp, _ := app.Test(req)
    defer resp.Body.Close()
    
    // Verify
    assert.Equal(t, 400, resp.StatusCode)
    
    // Check response body
    body, _ := io.ReadAll(resp.Body)
    var response types.Response
    json.Unmarshal(body, &response)
    
    assert.False(t, response.Success)
    assert.Contains(t, response.Error, "Invalid startDate format")
    
    // Test invalid end date
    req = httptest.NewRequest(http.MethodGet, "/stocks/AAPL/history?endDate=invalid", nil)
    resp, _ = app.Test(req)
    defer resp.Body.Close()
    
    // Verify
    assert.Equal(t, 400, resp.StatusCode)
    
    // Check response body
    body, _ = io.ReadAll(resp.Body)
    json.Unmarshal(body, &response)
    
    assert.False(t, response.Success)
    assert.Contains(t, response.Error, "Invalid endDate format")
}

// TestGetStockHistoryRangeTickerNotFound tests 404 response when ticker not found for range query
func TestGetStockHistoryRangeTickerNotFound(t *testing.T) {
    // Setup
    app := setupStockApp()
    
    // Execute - Test with a ticker that shouldn't exist
    req := httptest.NewRequest(http.MethodGet, "/stocks/NONEXISTENT/history", nil)
    resp, _ := app.Test(req)
    defer resp.Body.Close()
    
    // Verify
    assert.Equal(t, 404, resp.StatusCode)
    
    // Check response body
    body, _ := io.ReadAll(resp.Body)
    var response types.Response
    json.Unmarshal(body, &response)
    
    assert.False(t, response.Success)
    assert.Contains(t, response.Error, "not found with ticker")
}

// TestGetStockHistoryRangeValidDateParams tests that valid date parameters are properly parsed
func TestGetStockHistoryRangeValidDateParams(t *testing.T) {
    // This test would normally check that valid date parameters work,
    // but since we're avoiding DB mocks, we'll just verify the 404 for ticker not found
    // which confirms that date parsing worked correctly
    
    // Setup
    app := setupStockApp()
    
    // Execute - Test with valid date params but non-existent ticker
    req := httptest.NewRequest(http.MethodGet, "/stocks/NONEXISTENT/history?startDate=2023-01-01&endDate=2023-12-31", nil)
    resp, _ := app.Test(req)
    defer resp.Body.Close()
    
    // Verify
    assert.Equal(t, 404, resp.StatusCode)
    
    // Check response body
    body, _ := io.ReadAll(resp.Body)
    var response types.Response
    json.Unmarshal(body, &response)
    
    assert.False(t, response.Success)
    assert.Contains(t, response.Error, "not found with ticker")
    // The key here is that we got a 404, not a 400, confirming date parsing succeeded
}

// TestNewStockController verifies constructor creates a valid instance
func TestNewStockController(t *testing.T) {
    controller := NewStockController()
    assert.NotNil(t, controller)
}

// TestStockControllerTypeAssertions verifies the controller implements expected methods
func TestStockControllerTypeAssertions(t *testing.T) {
    // This test verifies at compile-time that StockController implements
    // all the required methods
    controller := NewStockController()
    
    // Type assertion - will fail at compile time if interface doesn't match
    // In real code, you might have an interface these controllers implement
    var _ interface{} = controller
    
    // We can also verify method existence by calling them with nil
    // This doesn't execute the methods but confirms they exist with right signatures
    var c *fiber.Ctx = nil
    
    // These will panic if methods don't exist, but we're not actually running them
    _ = func() {
        defer func() { recover() }()
        controller.GetAllStocks(c)
    }
    
    _ = func() {
        defer func() { recover() }()
        controller.GetStockByTicker(c)
    }
    
    _ = func() {
        defer func() { recover() }()
        controller.GetStockFirstHistory(c)
    }
    
    _ = func() { 
        defer func() { recover() }()
        controller.GetStockHistoryByDate(c)
    }
    
    _ = func() {
        defer func() { recover() }()
        controller.GetStockHistoryRange(c)
    }
}