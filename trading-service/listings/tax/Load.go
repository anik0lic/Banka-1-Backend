package tax

import (
	"log"
	"time"

	"banka1.com/db"
	"banka1.com/types"
)

func LoadTax() {

	monthYear := time.Now().Format("2006-01")

	var count int64
	db.DB.Model(&types.Tax{}).Where("user_id = ? AND month_year = ?", 3, monthYear).Count(&count)
	if count > 0 {
		log.Println("Tax already exists, skip adding.")
		return
	}

	taxData := types.Tax{
		UserID:        3,
		MonthYear:     monthYear,
		TaxableProfit: 50000.00,
		TaxAmount:     15000.00,
		IsPaid:        false,
		CreatedAt:     time.Now().Format("2006-01-02"),
	}

	if err := db.DB.Create(&taxData).Error; err != nil {
		log.Println("Failed to insert tax:", err)
		return
	}

	log.Println("Tax record inserted successfully!")
}
