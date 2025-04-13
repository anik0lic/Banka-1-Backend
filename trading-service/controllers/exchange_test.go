package controllers

import (
    "banka1.com/types"
    "encoding/json"
    "github.com/gofiber/fiber/v2"
    "github.com/stretchr/testify/assert"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func TestInitExchangeRoutes(t *testing.T) {
    // Setup
    app := fiber.New()
    
    // Execute
    InitExchangeRoutes(app)
    
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
    assert.True(t, findRoute("GET", "/exchanges"))
    assert.True(t, findRoute("GET", "/exchanges/:id"))
    assert.True(t, findRoute("GET", "/exchanges/mic/:micCode"))
    assert.True(t, findRoute("GET", "/exchanges/acronym/:acronym"))
}

func TestGetExchangeByIDInvalidID(t *testing.T) {
    // Setup
    app := fiber.New()
    controller := NewExchangeController()
    app.Get("/exchanges/:id", controller.GetExchangeByID)
    
    // Execute with invalid ID (non-existent)
    req := httptest.NewRequest(http.MethodGet, "/exchanges/999999", nil)
    resp, _ := app.Test(req)
    
    // Since we're not mocking the DB, this should return a 404
    assert.Equal(t, 404, resp.StatusCode)
}

func TestGetExchangeByMICInvalidMIC(t *testing.T) {
    // Setup
    app := fiber.New()
    controller := NewExchangeController()
    app.Get("/exchanges/mic/:micCode", controller.GetExchangeByMIC)
    
    // Execute with non-existent MIC
    req := httptest.NewRequest(http.MethodGet, "/exchanges/mic/INVALID", nil)
    resp, _ := app.Test(req)
    
    // Verify
    assert.Equal(t, 404, resp.StatusCode)
    
    // Check response body
    body, _ := io.ReadAll(resp.Body)
    var response types.Response
    json.Unmarshal(body, &response)
    
    assert.False(t, response.Success)
    assert.Contains(t, response.Error, "ExchangeMic not found with MIC code")
}

func TestGetExchangeByAcronymInvalidAcronym(t *testing.T) {
    // Setup
    app := fiber.New()
    controller := NewExchangeController()
    app.Get("/exchanges/acronym/:acronym", controller.GetExchangeByAcronym)
    
    // Execute with non-existent acronym
    req := httptest.NewRequest(http.MethodGet, "/exchanges/acronym/NONEXISTENT", nil)
    resp, _ := app.Test(req)
    
    // Verify
    assert.Equal(t, 404, resp.StatusCode)
    
    // Check response body
    body, _ := io.ReadAll(resp.Body)
    var response types.Response
    json.Unmarshal(body, &response)
    
    assert.False(t, response.Success)
    assert.Contains(t, response.Error, "ExchangeMic not found with acronym")
}

func TestNewExchangeController(t *testing.T) {
    // Verify that the constructor returns a non-nil controller
    controller := NewExchangeController()
    assert.NotNil(t, controller)
}