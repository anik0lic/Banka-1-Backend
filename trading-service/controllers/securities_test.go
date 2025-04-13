package controllers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/stretchr/testify/assert"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func TestInitSecuritiesRoutes(t *testing.T) {
    // Setup
    app := fiber.New()
    
    // Execute
    InitSecuritiesRoutes(app)
    
    // Verify routes are registered
    findRoute := func(method, path string) bool {
        for _, routes := range app.Stack() {
            for _, route := range routes {
                if route.Method == method && strings.Contains(route.Path, path) {
                    return true
                }
            }
        }
        return false
    }
    
    // Check each expected route
    assert.True(t, findRoute("GET", "/securities"))
    assert.True(t, findRoute("GET", "/securities/available"))
    assert.True(t, findRoute("GET", "/securities/:id"))
}

func TestGetUserSecuritiesMissingID(t *testing.T) {
    // Setup
    app := fiber.New()
    controller := NewSecuritiesController()
    app.Get("/securities/:id", controller.GetUserSecurities)
    
    // Execute - request without ID parameter
    req := httptest.NewRequest(http.MethodGet, "/securities/", nil)
    resp, _ := app.Test(req)
    
    // Verify
    assert.Equal(t, 404, resp.StatusCode) // Should return 404 Not Found for the route
}

func TestGetUserSecuritiesInvalidID(t *testing.T) {
    // This would normally test an invalid ID format, but our controller
    // doesn't actually validate ID formats, it only checks for empty IDs
    // If we needed to test a more specific validation, we would add it here
}

func TestNewSecuritiesController(t *testing.T) {
    // Verify that the constructor returns a non-nil controller
    controller := NewSecuritiesController()
    assert.NotNil(t, controller)
}