package main

import (
	"cashflow-backend/pkg/configs"
	"cashflow-backend/pkg/routes"
	"fmt"
	"log"
	"os"

	fiberSwagger "github.com/gofiber/swagger"

	_ "cashflow-backend/docs"

	"github.com/gofiber/fiber/v2"
)

// @title           Cashflow API Documentation
// @version         1.0
// @description     Dokumentasi API untuk aplikasi Cashflow.
// @termsOfService  http://swagger.io/terms/

// @contact.name    Bagas Ramadhan Rusnadi
// @contact.email   bagasramadhan239@gmail.com

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:3000
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token dengan format: "Bearer <token_jwt_disini>"

func main(){
	configs.ConnectDB()

	app := fiber.New()

	// Route Swagger
    // Nanti akses di: http://localhost:3000/swagger/index.html
	app.Get("/swagger/*", fiberSwagger.HandlerDefault)
	
	routes.SetupRoutes(app)
	
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}
	fmt.Println("Server is running on port", port)
	log.Fatal(app.Listen(port))
}