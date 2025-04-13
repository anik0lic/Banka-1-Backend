package cron

import (
    "banka1.com/db"
    "banka1.com/types"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    _ "github.com/mattn/go-sqlite3"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// MockDBWrapper helps us mock the database interactions
type MockDB struct {
    mock.Mock
}

func (m *MockDB) Find(dest interface{}, conds ...interface{}) *MockDB {
    args := m.Called(dest)
    return args.Get(0).(*MockDB)
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *MockDB {
    m.Called(query, args)
    return m
}

func (m *MockDB) First(dest interface{}) *MockDB {
    args := m.Called(dest)
    return args.Get(0).(*MockDB)
}

func (m *MockDB) Create(value interface{}) *MockDB {
    args := m.Called(value)
    return args.Get(0).(*MockDB)
}

func (m *MockDB) Save(value interface{}) *MockDB {
    args := m.Called(value)
    return args.Get(0).(*MockDB)
}

func (m *MockDB) Update(column string, value interface{}) *MockDB {
    args := m.Called(column, value)
    return args.Get(0).(*MockDB)
}

func (m *MockDB) Model(value interface{}) *MockDB {
    args := m.Called(value)
    return args.Get(0).(*MockDB)
}

func (m *MockDB) Error() error {
    args := m.Called()
    return args.Error(0)
}

// Setup an in-memory database for testing
func setupTestDB() *gorm.DB {
    // Use a file-based in-memory database with the pure Go driver
    testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        panic("Failed to open in-memory database: " + err.Error())
    }

    // Auto migrate the necessary models
    testDB.AutoMigrate(&types.Listing{})
    testDB.AutoMigrate(&types.ListingHistory{})
    testDB.AutoMigrate(&types.OptionContract{})
    testDB.AutoMigrate(&types.Actuary{})

    return testDB
}

// TestSnapshotListingsToHistory tests the SnapshotListingsToHistory function
func TestSnapshotListingsToHistory(t *testing.T) {
    // Setup
    originalDB := db.DB
    testDB := setupTestDB()
    db.DB = testDB
    defer func() { db.DB = originalDB }()

    // Create sample listings
    testListings := []types.Listing{
        {
            ID:           1,
            Ticker:       "AAPL",
            Name:         "Apple Inc.",
            ExchangeID:   1,
            LastRefresh:  time.Now(),
            Price:        150.50,
            Ask:          151.00,
            Bid:          150.00,
            Type:         "Stock",
            Subtype:      "Common Stock",
            ContractSize: 1,
        },
        {
            ID:           2,
            Ticker:       "MSFT",
            Name:         "Microsoft Corporation",
            ExchangeID:   1,
            LastRefresh:  time.Now(),
            Price:        300.25,
            Ask:          301.00,
            Bid:          300.00,
            Type:         "Stock",
            Subtype:      "Common Stock",
            ContractSize: 1,
        },
    }

    // Insert listings
    for _, listing := range testListings {
        testDB.Create(&listing)
    }

    // Execute
    err := SnapshotListingsToHistory()

    // Verify
    assert.NoError(t, err)

    // Check that histories were created
    var histories []types.ListingHistory
    testDB.Find(&histories)
    assert.Equal(t, 2, len(histories))
    
    // Verify specific details of histories
    assert.Equal(t, "AAPL", histories[0].Ticker)
    assert.Equal(t, "MSFT", histories[1].Ticker)
    assert.Equal(t, 150.50, histories[0].Price)
    assert.Equal(t, 300.25, histories[1].Price)
    
    // Test idempotency - running again shouldn't create duplicates
    err = SnapshotListingsToHistory()
    assert.NoError(t, err)
    
    var historiesAfterSecondRun []types.ListingHistory
    testDB.Find(&historiesAfterSecondRun)
    assert.Equal(t, 2, len(historiesAfterSecondRun))
}

// TestExpireOldOptionContracts tests the expireOldOptionContracts function
func TestExpireOldOptionContracts(t *testing.T) {
    // Setup
    originalDB := db.DB
    testDB := setupTestDB()
    db.DB = testDB
    defer func() { db.DB = originalDB }()

    // Create sample option contracts
    now := time.Now()
    oneWeekAgo := now.AddDate(0, 0, -7)
    oneWeekFuture := now.AddDate(0, 0, 7)
    
    testContracts := []types.OptionContract{
        {
            ID:           1,
            SettlementAt: oneWeekAgo, // Expired
            Status:       "active",
        },
        {
            ID:           2,
            SettlementAt: oneWeekFuture, // Not expired
            Status:       "active",
        },
        {
            ID:           3,
            SettlementAt: oneWeekAgo, // Already expired
            Status:       "expired",
        },
    }

    // Insert contracts
    for _, contract := range testContracts {
        testDB.Create(&contract)
    }

    // Execute
    expireOldOptionContracts()

    // Verify
    var contract1, contract2, contract3 types.OptionContract
    
    testDB.First(&contract1, 1)
    testDB.First(&contract2, 2)
    testDB.First(&contract3, 3)
    
    assert.Equal(t, "expired", contract1.Status, "Contract 1 should be expired")
    assert.Equal(t, "active", contract2.Status, "Contract 2 should still be active")
    assert.Equal(t, "expired", contract3.Status, "Contract 3 should remain expired")
}

// TestResetDailyLimits tests the resetDailyLimits function
func TestResetDailyLimits(t *testing.T) {
    // Setup
    originalDB := db.DB
    testDB := setupTestDB()
    db.DB = testDB
    defer func() { db.DB = originalDB }()

    // Create sample actuaries
    testActuaries := []types.Actuary{
        {
            ID:          1,
            UserID:      101,
            Department:  "Trading",
            FullName:    "John Agent",
            LimitAmount: 10000,
            UsedLimit:   5000,
        },
        {
            ID:          2,
            UserID:      102,
            Department:  "Trading",
            FullName:    "Jane Manager",
            LimitAmount: 50000,
            UsedLimit:   20000,
        },
    }

    // Insert actuaries
    for _, actuary := range testActuaries {
        testDB.Create(&actuary)
    }

    // Execute
    resetDailyLimits()

    // Verify
    var agent, manager types.Actuary
    
    testDB.First(&agent, 1)
    testDB.First(&manager, 2)
    
    assert.Equal(t, float64(0), agent.UsedLimit, "Agent's used limit should be reset to 0")
    assert.Equal(t, float64(20000), manager.UsedLimit, "Manager's used limit should not be reset")
}

// TestEmployeeToActuary tests the employeeToActuary function
func TestEmployeeToActuary(t *testing.T) {
    // Setup
    employee := Employee{
        ID:        123,
        FirstName: "Alice",
        LastName:  "Smith",
        Email:     "alice.smith@example.com",
        Department: "Risk",
        Position:   "Analyst",
        Active:     true,
        Permissions: []string{"trade", "view"},
    }

    // Execute
    actuary := employeeToActuary(employee)

    // Verify
    assert.Equal(t, uint(123), actuary.UserID)
    assert.Equal(t, "Risk", actuary.Department)
    assert.Equal(t, "Alice Smith", actuary.FullName)
    assert.Equal(t, "alice.smith@example.com", actuary.Email)
    assert.Equal(t, float64(100000), actuary.LimitAmount)
}

// TestGetActuaries tests the GetActuaries function
func TestGetActuaries(t *testing.T) {
    // Setup mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check that the request has the correct path
        assert.Equal(t, "/api/users/employees/actuaries", r.URL.Path)
        
        // Return a mock response
        response := APIResponse{
            Success: true,
            Data: []Employee{
                {
                    ID:        123,
                    FirstName: "John",
                    LastName:  "Doe",
                    Email:     "john.doe@example.com",
                    Department: "Trading",
                    Position:   "Analyst",
                    Active:     true,
                    Permissions: []string{"trade"},
                },
                {
                    ID:        456,
                    FirstName: "Jane",
                    LastName:  "Smith",
                    Email:     "jane.smith@example.com",
                    Department: "Risk",
                    Position:   "Manager",
                    Active:     true,
                    Permissions: []string{"trade", "approve"},
                },
            },
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()
    
    // Set environment variable to point to the test server
    originalUserService := os.Getenv("USER_SERVICE")
    os.Setenv("USER_SERVICE", server.URL)
    defer os.Setenv("USER_SERVICE", originalUserService)

    // Execute
    response, err := GetActuaries()

    // Verify
    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.True(t, response.Success)
    assert.Equal(t, 2, len(response.Data))
    assert.Equal(t, "John", response.Data[0].FirstName)
    assert.Equal(t, "Doe", response.Data[0].LastName)
    assert.Equal(t, "jane.smith@example.com", response.Data[1].Email)
}

// TestCreateNewActuaries tests the createNewActuaries function
func TestCreateNewActuaries(t *testing.T) {
    // Setup
    originalDB := db.DB
    testDB := setupTestDB()
    db.DB = testDB
    defer func() { db.DB = originalDB }()
    
    // Setup mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        response := APIResponse{
            Success: true,
            Data: []Employee{
                {
                    ID:        123,
                    FirstName: "John",
                    LastName:  "Doe",
                    Email:     "john.doe@example.com",
                    Department: "Trading",
                    Position:   "Analyst",
                    Active:     true,
                    Permissions: []string{"trade"},
                },
                {
                    ID:        456,
                    FirstName: "Jane",
                    LastName:  "Smith",
                    Email:     "jane.smith@example.com",
                    Department: "Risk",
                    Position:   "Manager",
                    Active:     true,
                    Permissions: []string{"trade", "approve"},
                },
            },
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()
    
    // Set environment variable
    originalUserService := os.Getenv("USER_SERVICE")
    os.Setenv("USER_SERVICE", server.URL)
    defer os.Setenv("USER_SERVICE", originalUserService)

    // Execute
    createNewActuaries()

    // Verify
    var actuaries []types.Actuary
    testDB.Find(&actuaries)
    
    assert.Equal(t, 2, len(actuaries))
    assert.Equal(t, uint(123), actuaries[0].UserID)
    assert.Equal(t, "John Doe", actuaries[0].FullName)
    assert.Equal(t, uint(456), actuaries[1].UserID)
    assert.Equal(t, "jane.smith@example.com", actuaries[1].Email)
    
    // Test idempotency - running again shouldn't create duplicates
    createNewActuaries()
    
    var actuariesAfterSecondRun []types.Actuary
    testDB.Find(&actuariesAfterSecondRun)
    assert.Equal(t, 2, len(actuariesAfterSecondRun))
}

// TestStartScheduler tests the StartScheduler function
func TestStartScheduler(t *testing.T) {
    // This is mainly a smoke test since we can't easily test the scheduled execution
    
    // Setup
    originalDB := db.DB
    testDB := setupTestDB()
    db.DB = testDB
    defer func() { db.DB = originalDB }()
    
    // Setup mock server for GetActuaries
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        response := APIResponse{
            Success: true,
            Data: []Employee{},
        }
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()
    
    originalUserService := os.Getenv("USER_SERVICE")
    os.Setenv("USER_SERVICE", server.URL)
    defer os.Setenv("USER_SERVICE", originalUserService)

    // Execute - this shouldn't panic
    assert.NotPanics(t, func() {
        StartScheduler()
    })
}