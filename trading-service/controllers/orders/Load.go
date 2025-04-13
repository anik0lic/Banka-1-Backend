package orders

import (
	"log"

	"banka1.com/db"
	"banka1.com/types"
)

func LoadOrders() {

	var msftSecurity types.Security
	var aaplSecurity types.Security
	var googlSecurity types.Security
	var nvdaSecurity types.Security

	db.DB.Where("ticker = ?", "MSFT").First(&msftSecurity)
	db.DB.Where("ticker = ?", "AAPL").First(&aaplSecurity)
	db.DB.Where("ticker = ?", "GOOGL").First(&googlSecurity)
	db.DB.Where("ticker = ?", "NVDA").First(&nvdaSecurity)

	order1 := types.Order{
		UserID:       3,
		AccountID:    1,
		SecurityID:   msftSecurity.ID,
		OrderType:    "Market",
		Quantity:     10,
		ContractSize: 1,
		Direction:    "Buy",
		Status:       "Approved",
		IsDone:       true,
	}

	order2 := types.Order{
		UserID:       3,
		AccountID:    1,
		SecurityID:   aaplSecurity.ID,
		OrderType:    "Market",
		Quantity:     5,
		ContractSize: 1,
		Direction:    "Buy",
		Status:       "Approved",
		IsDone:       true,
	}

	if err := db.DB.FirstOrCreate(&order1, types.Order{
		UserID:     order1.UserID,
		SecurityID: order1.SecurityID,
		Direction:  order1.Direction,
	}).Error; err != nil {
		log.Println("Error while adding order1:", err)
	} else {
		log.Printf("Order succesfully added to Security ID %d (order1)\n", order1.SecurityID)
	}

	if err := db.DB.FirstOrCreate(&order2, types.Order{
		UserID:     order2.UserID,
		SecurityID: order2.SecurityID,
		Direction:  order2.Direction,
	}).Error; err != nil {
		log.Println("Error while adding order2:", err)
	} else {
		log.Printf("Order succesfully added to Security ID %d (order2)\n", order2.SecurityID)
	}
}

func LoadPortfolios() {

	// get msft Security
	// select from securities where ticker = 'MSFT'

	var msftSecurity types.Security
	var aaplSecurity types.Security
	var googlSecurity types.Security
	var nvdaSecurity types.Security

	db.DB.Where("ticker = ?", "MSFT").First(&msftSecurity)
	db.DB.Where("ticker = ?", "AAPL").First(&aaplSecurity)
	db.DB.Where("ticker = ?", "GOOGL").First(&googlSecurity)
	db.DB.Where("ticker = ?", "NVDA").First(&nvdaSecurity)

	log.Println("MSFT Security preuzeta:", msftSecurity.Ticker)

	portfolio1 := types.Portfolio{
		UserID:        1,
		SecurityID:    aaplSecurity.ID, // AAPL
		Quantity:      20,
		PurchasePrice: 199.99,
		PublicCount:   10,
	}

	portfolio2 := types.Portfolio{
		UserID:        1,
		SecurityID:    googlSecurity.ID, // GOOGL
		Quantity:      20,
		PurchasePrice: 299.99,
		PublicCount:   10,
	}

	portfolio3 := types.Portfolio{
		UserID:        3,
		SecurityID:    msftSecurity.ID, // MSFT
		Quantity:      20,
		PurchasePrice: 299.99,
		PublicCount:   10,
	}

	portfolio4 := types.Portfolio{
		UserID:        3,
		SecurityID:    nvdaSecurity.ID, // NVDA
		Quantity:      20,
		PurchasePrice: 299.99,
		PublicCount:   10,
	}

	if err := db.DB.FirstOrCreate(&portfolio1, types.Portfolio{
		UserID:     portfolio1.UserID,
		SecurityID: portfolio1.SecurityID,
	}).Error; err != nil {
		log.Println("Greška pri dodavanju portfolio3:", err)
	} else {
		log.Println("Portfolio1 uspešno dodat")
	}
	if err := db.DB.FirstOrCreate(&portfolio2, types.Portfolio{
		UserID:     portfolio2.UserID,
		SecurityID: portfolio2.SecurityID,
	}).Error; err != nil {
		log.Println("Greška pri dodavanju portfolio4:", err)
	} else {
		log.Println("Portfolio2 uspešno dodat")
	}
	if err := db.DB.FirstOrCreate(&portfolio3, types.Portfolio{
		UserID:     portfolio3.UserID,
		SecurityID: portfolio3.SecurityID,
	}).Error; err != nil {
		log.Println("Greška pri dodavanju portfolio5:", err)
	} else {
		log.Println("Portfolio3 uspešno dodat")
	}
	if err := db.DB.FirstOrCreate(&portfolio4, types.Portfolio{
		UserID:     portfolio4.UserID,
		SecurityID: portfolio4.SecurityID,
	}).Error; err != nil {
		log.Println("Greška pri dodavanju portfolio5:", err)
	} else {
		log.Println("Portfolio4 uspešno dodat")
	}
}
