package controllers

import (
	"banka1.com/db"
	"banka1.com/middlewares"
	"banka1.com/types"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

type PortfolioController struct {
}

func NewPortfolioController() *PortfolioController { return &PortfolioController{} }

type UpdatePublicCountRequest struct {
	SecurityID  uint `json:"security_id"`
	PublicCount int  `json:"public"`
}

// UpdatePublicCount godoc
//
//	@Summary		Ažuriranje broja javno oglašenih hartija
//	@Description	Menja broj hartija koje su označene kao javne u portfoliju korisnika.
//	@Tags			Portfolio
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int									true	"User ID"
//	@Param			body	body	UpdatePublicCountRequest			true	"Podaci za ažuriranje"
//	@Success		200	{object}	types.Response{data=string}			"Uspešna izmena"
//	@Failure		400	{object}	types.Response						"Nedostaje user ID ili telo nije ispravno"
//	@Failure		500	{object}	types.Response						"Greška pri ažuriranju"
//	@Router			/securities/{id}/public-count [put]
func (sc *PortfolioController) UpdatePublicCount(c *fiber.Ctx) error {

	var req UpdatePublicCountRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(types.Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return c.Status(401).JSON(types.Response{
			Success: false,
			Error:   "Unauthorized: Missing user ID",
		})
	}

	var userID uint
	switch v := userIDRaw.(type) {
	case float64:
		userID = uint(v)
	case int:
		userID = uint(v)
	case uint:
		userID = v
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return c.Status(401).JSON(types.Response{
				Success: false,
				Error:   "Unauthorized: Invalid user ID format",
			})
		}
		userID = uint(parsed)
	default:
		return c.Status(401).JSON(types.Response{
			Success: false,
			Error:   "Unauthorized: Unknown user ID type",
		})
	}

	// Izmena u bazi
	if err := db.DB.Model(&types.Portfolio{}).
		Where("user_id = ? AND security_id = ?", userID, req.SecurityID).
		Update("public_count", req.PublicCount).Error; err != nil {
		return c.Status(500).JSON(types.Response{
			Success: false,
			Error:   "Failed to update public count: " + err.Error(),
		})
	}

	return c.JSON(types.Response{
		Success: true,
		Data:    fmt.Sprintf("Updated public count to %d", req.PublicCount),
	})
}

func (pc *PortfolioController) GetAllPortfolios(c *fiber.Ctx) error {
	var portfolios []types.Portfolio

	if err := db.DB.Preload("Security").Find(&portfolios).Error; err != nil {
		return c.Status(500).JSON(types.Response{
			Success: false,
			Error:   "Greška pri dohvatanju portfolija: " + err.Error(),
		})
	}

	return c.JSON(types.Response{
		Success: true,
		Data:    portfolios,
	})
}

func InitPortfolioRoutes(app *fiber.App) {
	portfolioController := NewPortfolioController()

	app.Put("/securities/public-count", middlewares.Auth, portfolioController.UpdatePublicCount)
	app.Get("/portfolios", portfolioController.GetAllPortfolios)

}
