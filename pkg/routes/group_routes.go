package routes

import (
	"cashflow-backend/app/controllers"
	"cashflow-backend/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func GroupRoutes(api fiber.Router){
	group := api.Group("/groups")
	
	group.Use(middleware.Protected())
	group.Post("/", controllers.CreateGroup)
	group.Get("/", controllers.GetAllGroups)
}