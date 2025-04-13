package orders

import (
	"banka1.com/middlewares"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"banka1.com/db"
	"banka1.com/types"
	"gorm.io/gorm"
)

var (
	securityLocks = make(map[uint]*sync.Mutex)
	locksMu       sync.Mutex
	orderLocks    = make(map[uint]*sync.Mutex)
	orderLocksMu  sync.Mutex
)

// Funkcija koja vraća uvek isti mutex po securityID
func getLock(securityID uint) *sync.Mutex {
	locksMu.Lock()
	defer locksMu.Unlock()

	if _, exists := securityLocks[securityID]; !exists {
		securityLocks[securityID] = &sync.Mutex{}
	}
	return securityLocks[securityID]
}

func getOrderLock(orderID uint) *sync.Mutex {
	orderLocksMu.Lock()
	defer orderLocksMu.Unlock()

	if _, exists := orderLocks[orderID]; !exists {
		orderLocks[orderID] = &sync.Mutex{}
	}
	return orderLocks[orderID]
}

func MatchOrder(order types.Order) {
	go func() {
		// Zaključavanje po ORDER ID – sprečava paralelno izvršavanje istog ordera
		orderLock := getOrderLock(order.ID)
		orderLock.Lock()
		defer orderLock.Unlock()

		if err := db.DB.First(&order, order.ID).Error; err != nil {
			fmt.Printf("Greska pri refetch ordera %d: %v\n", order.ID, err)
			return
		}

		if order.AON {
			if !CanExecuteAll(order) {
				fmt.Println("AON: Nema dovoljno za celokupan order")
				return
			}
		}

		if !canPreExecute(order) {
			fmt.Println("Nije ispunjen uslov za order")
			return
		}

		for order.RemainingParts != nil && *order.RemainingParts > 0 {
			tx := db.DB.Begin()

			// Ponovo proveri order iz baze unutar transakcije
			if err := tx.First(&order, order.ID).Error; err != nil {
				fmt.Printf("Order nije pronađen u transakciji: %v\n", err)
				tx.Rollback()
				break
			}

			quantityToExecute := 1
			if *order.RemainingParts < quantityToExecute {
				quantityToExecute = *order.RemainingParts
			}

			price := getOrderPrice(order)

			token, err := middlewares.NewOrderToken(order.Direction, order.UserID, order.AccountID, price)
			url := fmt.Sprintf("%s/orders/execute/%s", os.Getenv("BANKING_SERVICE"), token)

			hadError := false
			if err == nil {
				agent := fiber.Post(url)

				statusCode, _, errs := agent.Bytes()

				fmt.Printf("📡 BANKING-RESPONSE: status=%d, errors=%v\n", statusCode, errs)

				if len(errs) != 0 || statusCode != 200 {
					hadError = true
				}

			} else {
				hadError = true
			}

			if hadError {
				fmt.Printf("Nalog %v nije izvršen\n", order.ID)
				break
			} else {
				executePartial(order, quantityToExecute, price, tx)

				if order.RemainingParts == nil || *order.RemainingParts == 0 {
					order.IsDone = true
					order.Status = "done"
				}

				// SKIDANJE unita ako je kupovina (smanjuje se dostupnost hartija)
				if order.Direction == "buy" {
					var security types.Security
					if err := tx.First(&security, order.SecurityID).Error; err == nil {
						security.Volume -= int64(quantityToExecute)
						if security.Volume < 0 {
							security.Volume = 0
						}
						tx.Save(&security)
					}
				}

				// Ažuriraj order u bazi (unutar transakcije)
				if err := tx.Model(&types.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
					"remaining_parts": *order.RemainingParts,
					"is_done":         order.IsDone,
					"status":          order.Status,
				}).Error; err != nil {
					fmt.Printf("Greska pri upisu order statusa: %v\n", err)
					tx.Rollback()
					break
				}

				if err := tx.Commit().Error; err != nil {
					fmt.Printf("Nalog %v nije izvršen: %v\n", order.ID, err)
					tx.Rollback()
					break
				} else {
					fmt.Printf("Nalog %v izvršen\n", order.ID)

					// REFETCH ORDER IZ BAZE
					if err := db.DB.First(&order, order.ID).Error; err != nil {
						fmt.Printf("Greska pri refetch ordera %d nakon commit-a: %v\n", order.ID, err)
						break
					}

					if order.RemainingParts == nil || *order.RemainingParts <= 0 {
						fmt.Printf("Order %d je već izvršen ili nema više delova za obradu\n", order.ID)
						break
					}
				}
			}

			delay := calculateDelay(order)
			time.Sleep(delay)
		}
	}()
}

func getListingPrice(order types.Order) float64 {
	var security types.Security
	err := db.DB.First(&security, order.SecurityID).Error
	if err != nil {
		fmt.Printf("Security nije pronadjen za ID %d: %v\n", order.SecurityID, err)
		return -1.0
	}

	var listing types.Listing
	err = db.DB.Where("ticker = ?", security.Ticker).First(&listing).Error
	if err != nil {
		fmt.Printf("Listing nije pronadjen za Ticker %s: %v\n", security.Ticker, err)
		return -1.0
	}

	if order.Direction == "sell" {
		return float64(listing.Bid)
	} else {
		return float64(listing.Ask)
	}
}

func getOrderPrice(order types.Order) float64 {
	if strings.ToUpper(order.OrderType) == "MARKET" {
		var security types.Security
		db.DB.First(&security, order.SecurityID)
		return security.LastPrice
	}
	if order.StopPricePerUnit != nil {
		return *order.StopPricePerUnit
	}
	if order.LimitPricePerUnit != nil {
		return *order.LimitPricePerUnit
	}
	return 0.0
}

func executePartial(order types.Order, quantity int, price float64, tx *gorm.DB) {
	lock := getLock(order.SecurityID)
	lock.Lock()
	defer lock.Unlock()

	if order.Status == "done" || order.RemainingParts == nil || *order.RemainingParts <= 0 {
		fmt.Printf("Order %d je već završen ili nema remaining parts\n", order.ID)
		return
	}

	var match types.Order
	direction := "buy"
	if order.Direction == "buy" {
		direction = "sell"
	} else {
		direction = "buy"
	}
	tx.Where("security_id = ? AND direction = ? AND status = ? AND is_done = ?", order.SecurityID, direction, "approved", false).
		Order("last_modified").
		First(&match)

	if match.ID == 0 {
		fmt.Println("Nema dostupnog ordera za matchovanje")
		return
	}

	// Provere za matchovani order
	if match.UserID == order.UserID {
		fmt.Println("Preskočen self-match")
		return
	}

	if match.AON && (match.RemainingParts == nil || *match.RemainingParts < quantity) {
		fmt.Println("Matchovani order je AON i ne može da se izvrši u celosti")
		return
	}

	if match.Margin {
		var actuary types.Actuary
		if err := tx.Where("user_id = ?", match.UserID).First(&actuary).Error; err != nil || actuary.Department != "agent" {
			fmt.Println("Matchovani margin order nema validnog aktuara")
			return
		}
		initialMargin := price * float64(quantity) * 0.3 * 1.1
		if actuary.LimitAmount-actuary.UsedLimit < initialMargin {
			fmt.Println("Matchovani margin order nema dovoljno limita")
			return
		}
	}

	if match.UserID == order.UserID {
		fmt.Println("Preskočen self-match")
		return
	}

	if !canPreExecute(match) {
		fmt.Println("Preskočen match sa nedovoljnim uslovima")
		return
	}

	matchQuantity := quantity

	if match.RemainingParts != nil && *match.RemainingParts < quantity {
		matchQuantity = *match.RemainingParts
	}

	if matchQuantity <= 0 || order.RemainingParts == nil || *order.RemainingParts < matchQuantity {
		fmt.Printf("Nevalidan matchQuantity = %d za Order %d\n", matchQuantity, order.ID)
		return
	}

	txn := types.Transaction{
		OrderID:      order.ID,
		BuyerID:      getBuyerID(order, match),
		SellerID:     getSellerID(order, match),
		SecurityID:   order.SecurityID,
		Quantity:     matchQuantity,
		PricePerUnit: price,
		TotalPrice:   price * float64(matchQuantity),
	}
	if err := tx.Create(&txn).Error; err != nil {
		fmt.Printf("Greska pri kreiranju transakcije: %v\n", err)
		return
	}

	*order.RemainingParts -= matchQuantity
	*match.RemainingParts -= matchQuantity

	if *match.RemainingParts == 0 {
		match.IsDone = true
		match.Status = "done"
	}
	if *order.RemainingParts == 0 {
		order.IsDone = true
		order.Status = "done"
	}

	//tx.Save(&order)
	//tx.Save(&match)
	if err := tx.Save(&order).Error; err != nil {
		fmt.Printf("Greska pri save za order: %v\n", err)
		return
	}
	if err := tx.Save(&match).Error; err != nil {
		fmt.Printf("Greska pri save za match: %v\n", err)
		return
	}

	updatePortfolio(getBuyerID(order, match), order.SecurityID, matchQuantity, tx)
	updatePortfolio(getSellerID(order, match), order.SecurityID, -matchQuantity, tx)

	if order.Margin {
		var actuary types.Actuary
		if err := tx.Where("user_id = ?", order.UserID).First(&actuary).Error; err == nil {
			initialMargin := price * float64(matchQuantity) * 0.3 * 1.1
			actuary.UsedLimit += initialMargin
			tx.Save(&actuary)
		}
	}

	fmt.Printf("Match success: Order %d ↔ Order %d za %d @ %.2f\n", order.ID, match.ID, matchQuantity, price)
}

func updatePortfolio(userID uint, securityID uint, delta int, tx *gorm.DB) {
	var portfolio types.Portfolio
	err := tx.Where("user_id = ? AND security_id = ?", userID, securityID).First(&portfolio).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if delta > 0 {
			portfolio = types.Portfolio{
				UserID:     userID,
				SecurityID: securityID,
				Quantity:   delta,
			}
			if err := tx.Create(&portfolio).Error; err != nil {
				fmt.Printf("Portfolio greška u create: user=%d, security=%d, delta=%d | %v\n", userID, securityID, delta, err)
			} else {
				fmt.Printf("Portfolio kreiran: user=%d, security=%d, quantity=%d\n", userID, securityID, delta)
			}
		} else {
			fmt.Printf("Nema postojeći portfolio za korisnika %d i security %d, a pokušaj da se oduzme delta=%d\n", userID, securityID, delta)
		}
		return
	}

	if err != nil {
		fmt.Printf("Greska pri dohvatanju portfolia: user=%d, security=%d | %v\n", userID, securityID, err)
		return
	}

	portfolio.Quantity += delta
	if portfolio.Quantity <= 0 {
		err = tx.Delete(&portfolio).Error
		if err != nil {
			fmt.Printf("Portfolio greška pri brisanju: user=%d, security=%d | %v\n", userID, securityID, err)
		} else {
			fmt.Printf("Portfolio obrisan: user=%d, security=%d\n", userID, securityID)
		}
	} else {
		err = tx.Save(&portfolio).Error
		if err != nil {
			fmt.Printf("Portfolio greška pri update: user=%d, security=%d | %v\n", userID, securityID, err)
		} else {
			fmt.Printf("Portfolio ažuriran: user=%d, security=%d, quantity=%d\n", userID, securityID, portfolio.Quantity)
		}
	}
}

func calculateDelay(order types.Order) time.Duration {
	delaySeconds := rand.Intn(10) + 1
	if order.AfterHours {
		return time.Duration(delaySeconds+1800) * time.Second
	}
	return time.Duration(delaySeconds) * time.Second
}

func getExecutableParts(order types.Order) int {
	var matchingOrders []types.Order
	direction := "buy"
	if order.Direction == "buy" {
		direction = "sell"
	} else {
		direction = "buy"
	}

	db.DB.Where("security_id = ? AND direction = ? AND status = ? AND is_done = ?", order.SecurityID, direction, "approved", false).Find(&matchingOrders)
	totalAvailable := 0
	for _, o := range matchingOrders {
		if o.RemainingParts != nil {
			totalAvailable += *o.RemainingParts
		}
	}

	return totalAvailable
}

func CanExecuteAll(order types.Order) bool {
	return getExecutableParts(order) >= *order.RemainingParts
}

func CanExecuteAny(order types.Order) bool {
	return getExecutableParts(order) > 0
}

func canPreExecute(order types.Order) bool {
	price := getListingPrice(order)
	if price < 0 {
		return false
	}
	if strings.ToUpper(order.OrderType) == "LIMIT" {
		if order.Direction == "sell" {
			return price >= *order.LimitPricePerUnit
		} else {
			return price <= *order.LimitPricePerUnit
		}
	} else if strings.ToUpper(order.OrderType) == "STOP" {
		if order.Direction == "sell" {
			return price <= *order.StopPricePerUnit
		} else {
			return price >= *order.StopPricePerUnit
		}
	} else if strings.ToUpper(order.OrderType) == "STOP-LIMIT" {
		if order.Direction == "sell" {
			return price <= *order.StopPricePerUnit && price >= *order.LimitPricePerUnit
		} else {
			return price >= *order.StopPricePerUnit && price <= *order.LimitPricePerUnit
		}
	}
	return true
}

func getBuyerID(a, b types.Order) uint {
	if a.Direction == "buy" {
		return a.UserID
	}
	return b.UserID
}

func getSellerID(a, b types.Order) uint {
	if a.Direction == "sell" {
		return a.UserID
	}
	return b.UserID
}
