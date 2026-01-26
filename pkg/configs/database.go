package configs

import (
	"cashflow-backend/app/models"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
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
    
    // --- PERBAIKAN 1: Tambahkan Config DisableForeignKey... ---
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        DisableForeignKeyConstraintWhenMigrating: true, 
    })
	// DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

    if err != nil {
        log.Fatalf("Gagal koneksi ke Database! Error: %v", err)
    }

    log.Println("Database connected successfully")

    sqlDB, err := DB.DB()
    if err != nil {
        log.Fatal(err)
    }
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    log.Println("Running AutoMigrate...")

    // --- PERBAIKAN 2: Urutkan dari INDUK ke ANAK ---
    // User & Group dibuat duluan. GroupMember & Transaction dibuat belakangan.
    err = DB.AutoMigrate(
        &models.User{},         // Induk 1
        &models.Wallet{},       // Induk 2 (bisa punya user/group)
        &models.Group{},        // Induk 3
        &models.GroupMember{},  // Anak (Butuh User & Group)
        &models.Transaction{},  // Anak (Butuh User & Wallet)
    )

    if err != nil {
        log.Fatal("Migration failed: ", err)
    }
    
    log.Println("Migration Success!")
}