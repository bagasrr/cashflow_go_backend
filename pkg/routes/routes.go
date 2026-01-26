package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
    api := app.Group("/api")

    
	AuthRoutes(api)
    WalletRoutes(api)
    UserRoutes(api) 
    TransactionRoutes(api) 
    GroupRoutes(api)
    
    // Nanti kalau ada Wallet, tinggal tambah:
    // WalletRoutes(api)
}