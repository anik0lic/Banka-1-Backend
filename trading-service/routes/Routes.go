package routes

import (
	"banka1.com/middlewares"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {

	//actuaryController := controllers.NewActuaryController()

	// Sve rute vezane za aktuare su dostupne iskljucivo supervizorima

	fmt.Println("usli u setup")

	app.Group("/actuaries", middlewares.Auth, middlewares.DepartmentCheck("supervisor"))

	//app.Get("/actuaries/profits", actuaryController.GetActuaryProfits)
	//app.Get("/securities/:id", portfolioController.GetUserSecurities)

}
