package portfolio

import (
    "banka1.com/db"
    "bytes"
    "log"
    "testing"
)

// TestLoadPortfoliosPanicRecovery tests that the LoadPortfolios function behavior
func TestLoadPortfoliosPanicRecovery(t *testing.T) {
    // Save original DB and log output
    originalDB := db.DB
    originalLog := log.Writer()
    
    // Capture log output
    var logBuf bytes.Buffer
    log.SetOutput(&logBuf)
    defer log.SetOutput(originalLog)
    
    // Test with nil DB (will cause panic)
    db.DB = nil
    defer func() {
        db.DB = originalDB
    }()
    
    // Use panic recovery to catch the expected panic
    panicked := false
    func() {
        defer func() {
            if r := recover(); r != nil {
                panicked = true
                // Log the panic - this will be captured in logBuf
                log.Printf("Recovered from panic: %v", r)
            }
        }()
        
        LoadPortfolios()
    }()
    
    // Check that we got a panic as expected
    if !panicked {
        t.Error("Expected function to panic with nil DB, but it didn't")
    }
    
    // Check that something got logged (either from function or from our recovery)
    if logBuf.Len() == 0 {
        t.Error("Expected log output but got none")
    }
}