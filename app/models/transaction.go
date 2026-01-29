package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	TransactionID uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"transaction_id"`
	UserID 			uuid.UUID 	`gorm:"type:uuid; index" json:"user_id"`
	WalletID 		uuid.UUID 	`gorm:"type:uuid; index" json:"wallet_id"`
	
	Title 			string 		`json:"title"`
	Amount 			float64 	`json:"amount"`
	Description 	string 		`json:"description"`
	Date 			time.Time 	`json:"date"`

	CategoryID    uuid.UUID `gorm:"type:uuid;not null" json:"category_id"`
    Category      Category  `gorm:"foreignKey:CategoryID" json:"category"`

    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`
    DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}