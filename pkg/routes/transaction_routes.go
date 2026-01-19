package routes

import (
	"cashflow-backend/app/controllers"
	"cashflow-backend/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func TransactionRoutes(api fiber.Router) {
    // Bikin group lagi khusus transactions, misal: /api/transactions
    trx := api.Group("/transactions")

	trx.Use(middleware.Protected())

    trx.Post("/", controllers.CreateTransaction)

    trx.Get("/", controllers.GetTransactions)
    trx.Get("/:id/details", controllers.GetTransactionByID)
    trx.Get("/:id/wallet/transaction", controllers.GetWalletWithTransactions)

    trx.Patch("/:id/update", controllers.UpdateTransaction)
    trx.Patch("/:id", controllers.SoftDeleteTransaction)
}