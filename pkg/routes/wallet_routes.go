package routes

import (
	"cashflow-backend/app/controllers"

	"github.com/gofiber/fiber/v2"
)

func WalletRoutes(api fiber.Router){
	wr := api.Group("/wallets")

	wr.Get("/", controllers.GetAllWallets)
	wr.Patch("/:id", controllers.UpdateWallet)
	wr.Delete("/:id", controllers.DeleteWallet)
}
