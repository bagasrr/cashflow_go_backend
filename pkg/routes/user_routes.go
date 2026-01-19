package routes

import (
	"cashflow-backend/app/controllers"
	"cashflow-backend/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// Kita terima parameter 'api' yang merupakan Group router
func UserRoutes(api fiber.Router) {
    user := api.Group("/users")

    user.Post("/register", controllers.RegisterUser)
    user.Delete("/:id", controllers.DeleteUser)
	
	user.Use(middleware.Protected())
    user.Get("/", controllers.GetAllActiveUser) 
	user.Get("/profile", controllers.GetMyProfile)
    user.Patch("/:id/upgrade", controllers.UpgradeUserPlan)
    // user.Patch("/:id/role", controllers.ChangeUserToOwner)
    
}