package main

import (
	"cashflow-backend/pkg/configs"
	"cashflow-backend/pkg/routes"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main(){
	configs.ConnectDB()

	app := fiber.New()

	
	routes.SetupRoutes(app)
	
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}
	fmt.Println("Server is running on port", port)
	log.Fatal(app.Listen(port))
}