package types

import (
    "encoding/json"
    "testing"
    "time"
)

func TestTransactionTableName(t *testing.T) {
    transaction := Transaction{}
    if transaction.TableName() != "transactions" {
        t.Errorf("Expected table name to be 'transactions', got %s", transaction.TableName())
    }
}

func TestTransactionCreation(t *testing.T) {
    now := time.Now()
    transaction := Transaction{
        ID:           1,
        OrderID:      100,
        BuyerID:      200,
        SellerID:     300,
        SecurityID:   400,
        Quantity:     10,
        PricePerUnit: 25.50,
        TotalPrice:   255.0,
        CreatedAt:    now,
    }

    if transaction.ID != 1 {
        t.Errorf("Expected ID to be 1, got %d", transaction.ID)
    }
    if transaction.OrderID != 100 {
        t.Errorf("Expected OrderID to be 100, got %d", transaction.OrderID)
    }
    if transaction.TotalPrice != 255.0 {
        t.Errorf("Expected TotalPrice to be 255.0, got %f", transaction.TotalPrice)
    }
}

func TestTaxResponseMarshaling(t *testing.T) {
    taxResponse := TaxResponse{
        UserID:        1,
        TaxableProfit: 1000.0,
        TaxAmount:     150.0,
        IsPaid:        false,
        IsActuary:     true,
    }

    jsonData, err := json.Marshal(taxResponse)
    if err != nil {
        t.Fatalf("Failed to marshal TaxResponse: %v", err)
    }

    var unmarshaledResponse TaxResponse
    err = json.Unmarshal(jsonData, &unmarshaledResponse)
    if err != nil {
        t.Fatalf("Failed to unmarshal TaxResponse: %v", err)
    }

    if unmarshaledResponse.UserID != taxResponse.UserID ||
        unmarshaledResponse.TaxableProfit != taxResponse.TaxableProfit ||
        unmarshaledResponse.TaxAmount != taxResponse.TaxAmount ||
        unmarshaledResponse.IsPaid != taxResponse.IsPaid ||
        unmarshaledResponse.IsActuary != taxResponse.IsActuary {
        t.Errorf("Unmarshaled TaxResponse does not match original")
    }
}

func TestAggregatedTaxResponseMarshaling(t *testing.T) {
    aggregatedResponse := AggregatedTaxResponse{
        UserID:          1,
        PaidThisYear:    500.0,
        UnpaidThisMonth: 100.0,
        IsActuary:       false,
    }

    jsonData, err := json.Marshal(aggregatedResponse)
    if err != nil {
        t.Fatalf("Failed to marshal AggregatedTaxResponse: %v", err)
    }

    var unmarshaledResponse AggregatedTaxResponse
    err = json.Unmarshal(jsonData, &unmarshaledResponse)
    if err != nil {
        t.Fatalf("Failed to unmarshal AggregatedTaxResponse: %v", err)
    }

    if unmarshaledResponse.UserID != aggregatedResponse.UserID ||
        unmarshaledResponse.PaidThisYear != aggregatedResponse.PaidThisYear ||
        unmarshaledResponse.UnpaidThisMonth != aggregatedResponse.UnpaidThisMonth ||
        unmarshaledResponse.IsActuary != aggregatedResponse.IsActuary {
        t.Errorf("Unmarshaled AggregatedTaxResponse does not match original")
    }
}

func TestSecurityResponseMarshaling(t *testing.T) {
    securityResponse := SecurityResponse{
        Security:      "AAPL",
        Symbol:        "AAPL",
        Amount:        100,
        PurchasePrice: 150.25,
        Profit:        25.75,
        LastModified:  "2023-01-01",
        Public:        true,
    }

    jsonData, err := json.Marshal(securityResponse)
    if err != nil {
        t.Fatalf("Failed to marshal SecurityResponse: %v", err)
    }

    var unmarshaledResponse SecurityResponse
    err = json.Unmarshal(jsonData, &unmarshaledResponse)
    if err != nil {
        t.Fatalf("Failed to unmarshal SecurityResponse: %v", err)
    }

    if unmarshaledResponse.Security != securityResponse.Security ||
        unmarshaledResponse.Symbol != securityResponse.Symbol ||
        unmarshaledResponse.Amount != securityResponse.Amount ||
        unmarshaledResponse.PurchasePrice != securityResponse.PurchasePrice ||
        unmarshaledResponse.Profit != securityResponse.Profit ||
        unmarshaledResponse.LastModified != securityResponse.LastModified ||
        unmarshaledResponse.Public != securityResponse.Public {
        t.Errorf("Unmarshaled SecurityResponse does not match original")
    }
}

func TestResponseMarshaling(t *testing.T) {
    data := map[string]interface{}{
        "id":   1,
        "name": "Test",
    }
    response := Response{
        Success: true,
        Data:    data,
        Error:   "",
    }

    jsonData, err := json.Marshal(response)
    if err != nil {
        t.Fatalf("Failed to marshal Response: %v", err)
    }

    var unmarshaledResponse Response
    err = json.Unmarshal(jsonData, &unmarshaledResponse)
    if err != nil {
        t.Fatalf("Failed to unmarshal Response: %v", err)
    }

    if unmarshaledResponse.Success != response.Success ||
        unmarshaledResponse.Error != response.Error {
        t.Errorf("Unmarshaled Response does not match original")
    }
}

func TestOTCTransactionDTOs(t *testing.T) {
    // Test OTCTransactionInitiationDTO
    initiationDTO := OTCTransactionInitiationDTO{
        Uid:             "123-456-789",
        SellerAccountId: 100,
        BuyerAccountId:  200,
        Amount:          1000.50,
    }

    jsonData, err := json.Marshal(initiationDTO)
    if err != nil {
        t.Fatalf("Failed to marshal OTCTransactionInitiationDTO: %v", err)
    }

    var unmarshaledInitiation OTCTransactionInitiationDTO
    err = json.Unmarshal(jsonData, &unmarshaledInitiation)
    if err != nil {
        t.Fatalf("Failed to unmarshal OTCTransactionInitiationDTO: %v", err)
    }

    if unmarshaledInitiation.Uid != initiationDTO.Uid ||
        unmarshaledInitiation.SellerAccountId != initiationDTO.SellerAccountId ||
        unmarshaledInitiation.BuyerAccountId != initiationDTO.BuyerAccountId ||
        unmarshaledInitiation.Amount != initiationDTO.Amount {
        t.Errorf("Unmarshaled OTCTransactionInitiationDTO does not match original")
    }

    // Test OTCTransactionACKDTO
    ackDTO := OTCTransactionACKDTO{
        Uid:     "123-456-789",
        Failure: false,
        Message: "Transaction completed successfully",
    }

    jsonData, err = json.Marshal(ackDTO)
    if err != nil {
        t.Fatalf("Failed to marshal OTCTransactionACKDTO: %v", err)
    }

    var unmarshaledAck OTCTransactionACKDTO
    err = json.Unmarshal(jsonData, &unmarshaledAck)
    if err != nil {
        t.Fatalf("Failed to unmarshal OTCTransactionACKDTO: %v", err)
    }

    if unmarshaledAck.Uid != ackDTO.Uid ||
        unmarshaledAck.Failure != ackDTO.Failure ||
        unmarshaledAck.Message != ackDTO.Message {
        t.Errorf("Unmarshaled OTCTransactionACKDTO does not match original")
    }
}

func TestOrderResponseMarshaling(t *testing.T) {
    stopPrice := 150.0
    limitPrice := 160.0
    approvedBy := uint(42)
    remainingParts := 5

    orderResponse := OrderResponse{
        ID:                1,
        AccountID:         100,
        UserID:            200,
        SecurityID:        300,
        Quantity:          10,
        ContractSize:      1,
        StopPricePerUnit:  &stopPrice,
        LimitPricePerUnit: &limitPrice,
        Direction:         "buy",
        Status:            "pending",
        ApprovedBy:        &approvedBy,
        IsDone:            false,
        LastModified:      1617295800,
        RemainingParts:    &remainingParts,
        AfterHours:        false,
        AON:               true,
        Margin:            false,
    }

    jsonData, err := json.Marshal(orderResponse)
    if err != nil {
        t.Fatalf("Failed to marshal OrderResponse: %v", err)
    }

    var unmarshaledOrder OrderResponse
    err = json.Unmarshal(jsonData, &unmarshaledOrder)
    if err != nil {
        t.Fatalf("Failed to unmarshal OrderResponse: %v", err)
    }

    if unmarshaledOrder.ID != orderResponse.ID ||
        unmarshaledOrder.AccountID != orderResponse.AccountID ||
        unmarshaledOrder.Direction != orderResponse.Direction ||
        *unmarshaledOrder.StopPricePerUnit != *orderResponse.StopPricePerUnit ||
        *unmarshaledOrder.LimitPricePerUnit != *orderResponse.LimitPricePerUnit ||
        *unmarshaledOrder.ApprovedBy != *orderResponse.ApprovedBy ||
        *unmarshaledOrder.RemainingParts != *orderResponse.RemainingParts ||
        unmarshaledOrder.AON != orderResponse.AON {
        t.Errorf("Unmarshaled OrderResponse does not match original")
    }
}

func TestCreateOrderRequestMarshaling(t *testing.T) {
    stopPrice := 150.0
    limitPrice := 160.0

    createRequest := CreateOrderRequest{
        UserID:            1,
        AccountID:         100,
        SecurityID:        200,
        Quantity:          10,
        ContractSize:      1,
        StopPricePerUnit:  &stopPrice,
        LimitPricePerUnit: &limitPrice,
        Direction:         "buy",
        AON:               true,
        Margin:            false,
    }

    jsonData, err := json.Marshal(createRequest)
    if err != nil {
        t.Fatalf("Failed to marshal CreateOrderRequest: %v", err)
    }

    var unmarshaledRequest CreateOrderRequest
    err = json.Unmarshal(jsonData, &unmarshaledRequest)
    if err != nil {
        t.Fatalf("Failed to unmarshal CreateOrderRequest: %v", err)
    }

    if unmarshaledRequest.UserID != createRequest.UserID ||
        unmarshaledRequest.AccountID != createRequest.AccountID ||
        unmarshaledRequest.SecurityID != createRequest.SecurityID ||
        unmarshaledRequest.Quantity != createRequest.Quantity ||
        *unmarshaledRequest.StopPricePerUnit != *createRequest.StopPricePerUnit ||
        *unmarshaledRequest.LimitPricePerUnit != *createRequest.LimitPricePerUnit ||
        unmarshaledRequest.Direction != createRequest.Direction ||
        unmarshaledRequest.AON != createRequest.AON ||
        unmarshaledRequest.Margin != createRequest.Margin {
        t.Errorf("Unmarshaled CreateOrderRequest does not match original")
    }
}

func TestActuaryModelMarshaling(t *testing.T) {
    actuary := Actuary{
        ID:           1,
        UserID:       100,
        Department:   "Risk Management",
        FullName:     "John Doe",
        Email:        "john.doe@example.com",
        LimitAmount:  10000.0,
        UsedLimit:    5000.0,
        NeedApproval: true,
    }

    jsonData, err := json.Marshal(actuary)
    if err != nil {
        t.Fatalf("Failed to marshal Actuary: %v", err)
    }

    var unmarshaledActuary Actuary
    err = json.Unmarshal(jsonData, &unmarshaledActuary)
    if err != nil {
        t.Fatalf("Failed to unmarshal Actuary: %v", err)
    }

    if unmarshaledActuary.ID != actuary.ID ||
        unmarshaledActuary.UserID != actuary.UserID ||
        unmarshaledActuary.Department != actuary.Department ||
        unmarshaledActuary.FullName != actuary.FullName ||
        unmarshaledActuary.Email != actuary.Email ||
        unmarshaledActuary.LimitAmount != actuary.LimitAmount ||
        unmarshaledActuary.UsedLimit != actuary.UsedLimit ||
        unmarshaledActuary.NeedApproval != actuary.NeedApproval {
        t.Errorf("Unmarshaled Actuary does not match original")
    }
}

func TestListingMarshaling(t *testing.T) {
    now := time.Now()
    
    listing := Listing{
        ID:           1,
        Ticker:       "AAPL",
        Name:         "Apple Inc.",
        ExchangeID:   1,
        Exchange:     Exchange{ID: 1, Name: "NASDAQ"},
        LastRefresh:  now,
        Price:        150.25,
        Ask:          150.50,
        Bid:          150.00,
        Type:         "Stock",
        Subtype:      "Common Stock",
        ContractSize: 1,
        CreatedAt:    now,
        UpdatedAt:    now,
    }

    jsonData, err := json.Marshal(listing)
    if err != nil {
        t.Fatalf("Failed to marshal Listing: %v", err)
    }

    var unmarshaledListing Listing
    err = json.Unmarshal(jsonData, &unmarshaledListing)
    if err != nil {
        t.Fatalf("Failed to unmarshal Listing: %v", err)
    }

    if unmarshaledListing.ID != listing.ID ||
        unmarshaledListing.Ticker != listing.Ticker ||
        unmarshaledListing.Name != listing.Name ||
        unmarshaledListing.ExchangeID != listing.ExchangeID ||
        unmarshaledListing.Price != listing.Price ||
        unmarshaledListing.Type != listing.Type ||
        unmarshaledListing.Subtype != listing.Subtype ||
        unmarshaledListing.ContractSize != listing.ContractSize {
        t.Errorf("Unmarshaled Listing does not match original")
    }
}

// Test for option-related types
func TestOptionCreation(t *testing.T) {
    now := time.Now()
    
    option := Option{
        ID:             1,
        ListingID:      100,
        OptionType:     "Call",
        StrikePrice:    150.0,
        ImpliedVol:     0.25,
        OpenInterest:   1000,
        SettlementDate: now.AddDate(0, 3, 0), // 3 months in the future
        ContractSize:   100,
        CreatedAt:      now,
        UpdatedAt:      now,
    }

    if option.ID != 1 {
        t.Errorf("Expected ID to be 1, got %d", option.ID)
    }
    
    if option.OptionType != "Call" {
        t.Errorf("Expected OptionType to be 'Call', got %s", option.OptionType)
    }
    
    if option.ContractSize != 100 {
        t.Errorf("Expected ContractSize to be 100, got %d", option.ContractSize)
    }
}

// Test for portfolio
func TestPortfolioModelRelationships(t *testing.T) {
    portfolio := Portfolio{
        ID:            1,
        UserID:        100,
        SecurityID:    200,
        Quantity:      10,
        PurchasePrice: 150.25,
        PublicCount:   5,
        CreatedAt:     time.Now().Unix(),
        Security: Security{
            ID:     200,
            UserID: 100,
            Ticker: "AAPL",
            Name:   "Apple Inc.",
        },
    }

    if portfolio.Security.ID != portfolio.SecurityID {
        t.Errorf("Portfolio's SecurityID (%d) doesn't match Security's ID (%d)", 
            portfolio.SecurityID, portfolio.Security.ID)
    }
}

// Test for tax
func TestTaxModelFields(t *testing.T) {
    tax := Tax{
        ID:            1,
        UserID:        100,
        MonthYear:     "2023-04",
        TaxableProfit: 1000.0,
        TaxAmount:     150.0,
        IsPaid:        false,
        CreatedAt:     "2023-04-12T15:04:05Z",
    }

    if tax.ID != 1 {
        t.Errorf("Expected ID to be 1, got %d", tax.ID)
    }
    
    if tax.UserID != 100 {
        t.Errorf("Expected UserID to be 100, got %d", tax.UserID)
    }
    
    if tax.MonthYear != "2023-04" {
        t.Errorf("Expected MonthYear to be '2023-04', got %s", tax.MonthYear)
    }
    
    if tax.TaxableProfit != 1000.0 {
        t.Errorf("Expected TaxableProfit to be 1000.0, got %f", tax.TaxableProfit)
    }
}