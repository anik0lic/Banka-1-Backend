package controllers

import (
	"banka1.com/controllers/orders"
	"banka1.com/db"
	"banka1.com/middlewares"
	"banka1.com/types"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"os"
	"strings"
)

type OrderController struct {
}

func NewOrderController() *OrderController {
	return &OrderController{}
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func OrderToOrderResponse(order types.Order) types.OrderResponse {
	return types.OrderResponse{
		ID:                order.ID,
		UserID:            order.UserID,
		AccountID:         order.AccountID,
		SecurityID:        order.SecurityID,
		Quantity:          order.Quantity,
		ContractSize:      order.ContractSize,
		StopPricePerUnit:  order.StopPricePerUnit,
		LimitPricePerUnit: order.LimitPricePerUnit,
		Direction:         order.Direction,
		Status:            order.Status,
		ApprovedBy:        order.ApprovedBy,
		IsDone:            order.IsDone,
		LastModified:      order.LastModified,
		RemainingParts:    order.RemainingParts,
		AfterHours:        order.AfterHours,
		AON:               order.AON,
		Margin:            order.Margin,
	}
}

// GetOrderByID godoc
//
//	@Summary		Preuzimanje naloga po I
//	@Summary		Preuzimanje naloga po ID-u
//	@Description	Vraća detalje specifičnog naloga na osnovu njegovog jedinstvenog identifikatora (ID).
//	@Tags			Orders
//	@Produce		json
//	@Param			id	path		int											true	"ID naloga koji se preuzima"
//	@Success		200	{object}	types.Response{data=types.OrderResponse}	"Uspešno preuzet nalog"
//	@Failure		400	{object}	types.Response								"Nevalidan ID naloga"
//	@Failure		404	{object}	types.Response								"Nalog sa datim ID-jem ne postoji"
//	@Router			/orders/{id} [get]
func (oc *OrderController) GetOrderByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id", -1)
	if err != nil || id <= 0 {
		var response types.Response
		response.Success = false
		if err != nil {
			response.Error = "Nevalidan ID: " + err.Error()
		} else {
			response.Error = "Nevalidan ID"
		}
		return c.Status(400).JSON(response)
	}
	var order types.Order
	if err := db.DB.First(&order, id).Error; err != nil {
		return c.Status(404).JSON(types.Response{
			Success: false,
			Error:   "Nije pronadjen: " + err.Error(),
		})
	}
	return c.JSON(types.Response{
		Success: true,
		Data:    OrderToOrderResponse(order),
	})
}

// GetOrders godoc
//
//	@Summary		Preuzimanje liste naloga
//	@Description	Vraća listu naloga, opciono filtriranu po statusu.
//	@Tags			Orders
//	@Produce		json
//	@Param			filter_status	query		string										false	"Status naloga za filtriranje. Podrazumevano 'all' za sve statuse."	default(all)	example(pending)
//	@Success		200				{object}	types.Response{data=[]types.OrderResponse}	"Uspešno preuzeta lista naloga"
//	@Failure		500				{object}	types.Response								"Greška pri preuzimanju naloga iz baze"
//	@Router			/orders [get]
func (oc *OrderController) GetOrders(c *fiber.Ctx) error {
	filterStatus := strings.ToLower(c.Query("filter_status", "all"))
	var ordersList []types.Order
	var err error
	if "all" == filterStatus {
		err = db.DB.Find(&ordersList).Error
	} else {
		err = db.DB.Find(&ordersList, "status = ?", filterStatus).Error
	}
	if err != nil {
		return c.Status(400).JSON(types.Response{
			Success: false,
			Error:   "Neuspela pretraga: " + err.Error(),
		})
	}
	responses := make([]types.OrderResponse, len(ordersList))
	for i, order := range ordersList {
		responses[i] = OrderToOrderResponse(order)
	}
	return c.JSON(types.Response{
		Success: true,
		Data:    responses,
	})
}

// CreateOrder godoc
//
//	@Summary		Kreiranje novog naloga
//	@Description	Kreira novi nalog za hartije od vrednosti.
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Param			orderRequest	body	types.CreateOrderRequest	true	"Podaci neophodni za kreiranje naloga"
//	@Security		BearerAuth
//	@Success		201	{object}	types.Response{data=uint}	"Uspešno kreiran nalog, vraća ID novog naloga"
//	@Failure		400	{object}	types.Response				"Neispravan format, neuspela validacija ili greška pri upisu u bazu"
//	@Failure		403	{object}	types.Response				"Nije dozvoljeno kreirati nalog za drugog korisnika"
//	@Router			/orders [post]
func (oc *OrderController) CreateOrder(c *fiber.Ctx) error {
	var orderRequest types.CreateOrderRequest
	userId := c.Locals("user_id").(float64)

	if err := c.BodyParser(&orderRequest); err != nil {
		return c.Status(400).JSON(types.Response{
			Success: false,
			Error:   "Neuspelo parsiranje: " + err.Error(),
		})
	}
	if err := validate.Struct(orderRequest); err != nil {
		return c.Status(400).JSON(types.Response{
			Success: false,
			Error:   "Neuspela validacija: " + err.Error(),
		})
	}
	if userId != float64(orderRequest.UserID) {
		return c.Status(403).JSON(types.Response{
			Success: false,
			Error:   "Cannot create order for another user",
		})
	}

	status := "pending"
	var approvedBy *uint = nil

	if deptRaw := c.Locals("department"); deptRaw != nil {
		if department, ok := deptRaw.(string); ok && department == "SUPERVISOR" {
			status = "approved"
			id := uint(userId)
			approvedBy = &id
		}
	}

	// Provera dostupnosti unita ako se order odobrava odmah
	if status == "approved" {
		var security types.Security
		if err := db.DB.First(&security, orderRequest.SecurityID).Error; err != nil {
			return c.Status(404).JSON(types.Response{
				Success: false,
				Error:   "Hartija nije pronađena",
			})
		}

		if strings.ToLower(orderRequest.Direction) == "buy" {
			if orderRequest.AON {
				if orderRequest.Quantity > int(security.Volume) {
					return c.Status(400).JSON(types.Response{
						Success: false,
						Error:   fmt.Sprintf("Nedovoljno dostupnih unita (%d dostupno)", security.Volume),
					})
				}
			}
		} else if strings.ToLower(orderRequest.Direction) == "sell" {
			var portfolio types.Portfolio
			if err := db.DB.Where("user_id = ? AND security_id = ?", orderRequest.UserID, orderRequest.SecurityID).First(&portfolio).Error; err != nil {
				return c.Status(400).JSON(types.Response{
					Success: false,
					Error:   "Nemate ovu hartiju u portfoliju",
				})
			}
			if orderRequest.AON && portfolio.Quantity < orderRequest.Quantity {
				return c.Status(400).JSON(types.Response{
					Success: false,
					Error:   fmt.Sprintf("Nemate dovoljno hartija za AON prodaju (imate %d, traženo %d)", portfolio.Quantity, orderRequest.Quantity),
				})
			}
		}
	}

	var orderType string
	switch {
	case orderRequest.StopPricePerUnit == nil && orderRequest.LimitPricePerUnit == nil:
		orderType = "MARKET"
	case orderRequest.StopPricePerUnit == nil && orderRequest.LimitPricePerUnit != nil:
		orderType = "LIMIT"
	case orderRequest.StopPricePerUnit != nil && orderRequest.LimitPricePerUnit == nil:
		orderType = "STOP"
	case orderRequest.StopPricePerUnit != nil && orderRequest.LimitPricePerUnit != nil:
		orderType = "STOP-LIMIT"
	}

	order := types.Order{
		UserID:            orderRequest.UserID,
		AccountID:         orderRequest.AccountID,
		SecurityID:        orderRequest.SecurityID,
		Quantity:          orderRequest.Quantity,
		ContractSize:      orderRequest.ContractSize,
		StopPricePerUnit:  orderRequest.StopPricePerUnit,
		LimitPricePerUnit: orderRequest.LimitPricePerUnit,
		OrderType:         orderType,
		Direction:         orderRequest.Direction,
		Status:            status, // TODO: pribaviti needs approval vrednost preko token-a?
		ApprovedBy:        approvedBy,
		IsDone:            false,
		RemainingParts:    &orderRequest.Quantity,
		AfterHours:        false, // TODO: dodati check za ovo
		AON:               orderRequest.AON,
		Margin:            orderRequest.Margin,
	}

	tx := db.DB.Create(&order)
	if err := tx.Error; err != nil {
		return c.Status(400).JSON(types.Response{
			Success: false,
			Error:   "Neuspelo kreiranje: " + err.Error(),
		})
	}

	if orderRequest.Margin {
		var security types.Security
		if err := db.DB.First(&security, orderRequest.SecurityID).Error; err != nil {
			return c.Status(404).JSON(types.Response{
				Success: false,
				Error:   "Hartija nije pronađena",
			})
		}

		maintenanceMargin := security.LastPrice * 0.3
		initialMarginCost := maintenanceMargin * 1.1

		department, hasDepartment := c.Locals("department").(string)

		if hasDepartment && (department == "AGENT" || department == "SUPERVISOR") {
			var actuary types.Actuary
			if err := db.DB.Where("user_id = ?", orderRequest.UserID).First(&actuary).Error; err != nil {
				return c.Status(403).JSON(types.Response{
					Success: false,
					Error:   "Korisnik nema margin nalog (nije agent ili nije registrovan kao aktuar)",
				})
			}

			if actuary.LimitAmount < initialMarginCost {
				return c.Status(403).JSON(types.Response{
					Success: false,
					Error:   "Nedovoljan limit za margin order",
				})
			}
		} else {
			client := &http.Client{}
			req, _ := http.NewRequest("GET", fmt.Sprintf("%s/loans/has-approved-loan/%d", os.Getenv("BANKING_SERVICE"), orderRequest.UserID), nil)
			//url := fmt.Sprintf("%s/orders/execute/%s", os.Getenv("BANKING_SERVICE"), token)

			req.Header = http.Header{
				"Authorization": []string{c.Get("Authorization")},
			}
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode != 200 {
				return c.Status(500).JSON(types.Response{
					Success: false,
					Error:   "Greška pri proveri kredita iz banking servisa",
				})
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				return c.Status(500).JSON(types.Response{
					Success: false,
					Error:   "Neuspešno parsiranje odgovora iz banking servisa",
				})
			}

			approved, ok := body["approvedLoan"].(bool)
			if !ok || !approved {
				return c.Status(403).JSON(types.Response{
					Success: false,
					Error:   "Korisnik nema prava za margin order (nema kredit ni permisiju)",
				})
			}
		}
	}

	return c.JSON(types.Response{
		Success: true,
		Data:    order.ID,
	})
}

func ApproveDeclineOrder(c *fiber.Ctx, decline bool) error {
	id, err := c.ParamsInt("id", -1)
	if err != nil || id <= 0 {
		var response types.Response
		response.Success = false
		if err != nil {
			response.Error = "Nevalidan ID: " + err.Error()
		} else {
			response.Error = "Nevalidan ID"
		}
		return c.Status(400).JSON(response)
	}
	var order types.Order
	if err := db.DB.First(&order, id).Error; err != nil {
		return c.Status(404).JSON(types.Response{
			Success: false,
			Error:   "Nije pronadjen: " + err.Error(),
		})
	}
	if order.Status != "pending" {
		return c.Status(400).JSON(types.Response{
			Success: false,
			Error:   "Nije na cekanju",
		})
	}
	if decline {
		order.Status = "declined"
	} else {
		// Provera dostupnosti unita pre odobrenja
		var security types.Security
		if err := db.DB.First(&security, order.SecurityID).Error; err != nil {
			return c.Status(404).JSON(types.Response{
				Success: false,
				Error:   "Hartija nije pronađena",
			})
		}

		if strings.ToLower(order.Direction) == "buy" {
			// Provera da li ima dostupnih unita za kupovinu
			if order.Quantity > int(security.Volume) {
				return c.Status(400).JSON(types.Response{
					Success: false,
					Error:   fmt.Sprintf("Nedovoljno dostupnih unita (%d dostupno)", security.Volume),
				})
			}
		} else if strings.ToLower(order.Direction) == "sell" {
			// Provera da li korisnik ima dovoljno hartija u portfoliju
			var portfolio types.Portfolio
			if err := db.DB.Where("user_id = ? AND security_id = ?", order.UserID, order.SecurityID).First(&portfolio).Error; err != nil {
				return c.Status(400).JSON(types.Response{
					Success: false,
					Error:   "Nemate ovu hartiju u portfoliju",
				})
			}
			if portfolio.Quantity < order.Quantity {
				return c.Status(400).JSON(types.Response{
					Success: false,
					Error:   fmt.Sprintf("Nemate dovoljno hartija da biste prodali (imate %d, traženo %d)", portfolio.Quantity, order.Quantity),
				})
			}
		}

		order.Status = "approved"
		order.ApprovedBy = new(uint)
		*order.ApprovedBy = 0
		db.DB.Save(&order)

		orders.MatchOrder(order)

		return c.JSON(types.Response{
			Success: true,
			Data:    fmt.Sprintf("Order %d odobren i pokrenuto izvršavanje", order.ID),
		})
	}

	order.ApprovedBy = new(uint)
	*order.ApprovedBy = 0 // TODO: dobavi iz token-a
	db.DB.Save(&order)

	return c.JSON(types.Response{
		Success: true,
		Data:    order.ID,
	})
}

// DeclineOrder godoc
//
//	@Summary		Odbijanje naloga
//	@Description	Menja status naloga u 'declined'.
//	@Tags			Orders
//	@Produce		json
//	@Param			id	path	int	true	"ID naloga koji se odbija"
//	@Security		BearerAuth
//	@Success		200	{object}	types.Response{data=uint}	"Nalog uspešno odbijen, vraća ID naloga"
//	@Failure		400	{object}	types.Response				"Nevalidan ID ili nalog nije u 'pending' statusu"
//	@Failure		403	{object}	types.Response				"Nedovoljne privilegije"
//	@Failure		404	{object}	types.Response				"Nalog sa datim ID-jem ne postoji"
//	@Failure		500	{object}	types.Response				"Interna Greška Servera"
//	@Router			/orders/{id}/decline [post]
func (oc *OrderController) DeclineOrder(c *fiber.Ctx) error {
	return ApproveDeclineOrder(c, true)
}

// ApproveOrder godoc
//
//	@Summary		Odobravanje naloga
//	@Description	Menja status naloga u 'approved'.
//	@Tags			Orders
//	@Produce		json
//	@Param			id	path	int	true	"ID naloga koji se odobrava"
//	@Security		BearerAuth
//	@Success		200	{object}	types.Response{data=uint}	"Nalog uspešno odobren"
//	@Failure		400	{object}	types.Response				"Nevalidan ID ili nalog nije u 'pending' statusu"
//	@Failure		403	{object}	types.Response				"Nedovoljne privilegije"
//	@Failure		404	{object}	types.Response				"Nalog sa datim ID-jem ne postoji"
//	@Failure		500	{object}	types.Response				"Interna Greška Servera"
//	@Router			/orders/{id}/approve [post]
func (oc *OrderController) ApproveOrder(c *fiber.Ctx) error {
	return ApproveDeclineOrder(c, false)
}

func InitOrderRoutes(app *fiber.App) {
	orderController := NewOrderController()

	app.Get("/orders/:id", orderController.GetOrderByID)
	app.Get("/orders", orderController.GetOrders)
	app.Post("/orders", middlewares.Auth, orderController.CreateOrder)
	app.Post("/orders/:id/decline", middlewares.Auth, middlewares.DepartmentCheck("SUPERVISOR"), orderController.DeclineOrder)
	app.Post("/orders/:id/approve", middlewares.Auth, middlewares.DepartmentCheck("SUPERVISOR"), orderController.ApproveOrder)
}
