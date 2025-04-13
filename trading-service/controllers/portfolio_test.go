package controllers

import (
    "banka1.com/types"
    "bytes"
    "encoding/json"
    "github.com/gofiber/fiber/v2"
    "github.com/stretchr/testify/assert"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func TestInitPortfolioRoutes(t *testing.T) {
    // Setup
    app := fiber.New()
    
    // Execute
    InitPortfolioRoutes(app)
    
    // Verify routes are registered
    findRoute := func(method, path string) bool {
        for methodRoutes := range app.Stack() {
            for _, route := range app.Stack()[methodRoutes] {
                if route.Method == method && strings.Contains(route.Path, path) {
                    return true
                }
            }
        }
        return false
    }
    
    // Check expected routes
    assert.True(t, findRoute("PUT", "/securities/public-count"))
    assert.True(t, findRoute("GET", "/portfolios"))
}

func TestUpdatePublicCountMissingUserID(t *testing.T) {
    // Setup
    app := fiber.New()
    controller := NewPortfolioController()
    app.Put("/securities/public-count", controller.UpdatePublicCount)
    
    // Create test request with valid body but missing user_id
    reqBody := `{"security_id": 1, "public": 5}`
    req := httptest.NewRequest(http.MethodPut, "/securities/public-count", bytes.NewBufferString(reqBody))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    
    // Verify
    assert.Equal(t, 401, resp.StatusCode)
    
    // Check response body
    body, _ := io.ReadAll(resp.Body)
    var response types.Response
    json.Unmarshal(body, &response)
    
    assert.False(t, response.Success)
    assert.Contains(t, response.Error, "Unauthorized")
}

func TestUpdatePublicCountInvalidRequestBody(t *testing.T) {
    // Setup
    app := fiber.New()
    controller := NewPortfolioController()
    app.Put("/securities/public-count", controller.UpdatePublicCount)
    
    // Create test request with invalid JSON
    reqBody := `{"security_id": "invalid", "public": true}`
    req := httptest.NewRequest(http.MethodPut, "/securities/public-count", bytes.NewBufferString(reqBody))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    
    // Verify
    assert.Equal(t, 400, resp.StatusCode)
    
    // Check response body
    body, _ := io.ReadAll(resp.Body)
    var response types.Response
    json.Unmarshal(body, &response)
    
    assert.False(t, response.Success)
    assert.Contains(t, response.Error, "Invalid request body")
}

func TestNewPortfolioController(t *testing.T) {
    // Verify that the constructor returns a non-nil controller
    controller := NewPortfolioController()
    assert.NotNil(t, controller)
}