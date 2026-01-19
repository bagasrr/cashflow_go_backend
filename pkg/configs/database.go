package configs

import (
	"cashflow-backend/app/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


var DB *gorm.DB

func ConnectDB(){
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_TIMEZONE"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Gagal koneksi ke Database! Error: %v", err)
	}

	log.Println("Database connected successfully")

	// DB.AutoMigrate(&models.Transaction{})
	// Pastikan urutan migrasinya benar (Induk dulu baru Anak)
	DB.AutoMigrate(&models.User{}, &models.Wallet{}, &models.Transaction{})
}