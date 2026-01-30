package routes

import (
	"cashflow-backend/app/controllers"
	"cashflow-backend/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func GroupRoutes(api fiber.Router){
	group := api.Group("/groups")
	
	group.Use(middleware.Protected())
	group.Get("/", controllers.GetAllGroups)
	group.Post("/", controllers.CreateGroup)
	group.Post("/add-new-member", controllers.AddPeopleToGroupByEmail)
}