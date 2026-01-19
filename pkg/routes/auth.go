package routes

import (
	"cashflow-backend/app/controllers"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(api fiber.Router){
	auth := api.Group("/auth")
	auth.Post("/login", controllers.Login)
}