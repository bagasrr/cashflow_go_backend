package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
    // 1. Buat Group Utama (/api)
    api := app.Group("/api")

    // 2. Panggil rute dari file sebelah
    // Kita passing variable 'api' ke fungsi mereka
	AuthRoutes(api)
    WalletRoutes(api)
    UserRoutes(api)        // Otomatis jadi /api/users/...
    TransactionRoutes(api) // Otomatis jadi /api/transactions/...
    
    // Nanti kalau ada Wallet, tinggal tambah:
    // WalletRoutes(api)
}