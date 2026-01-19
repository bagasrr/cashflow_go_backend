package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	TransactionID uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"transaction_id"`
    
    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`
    DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	UserID 			uuid.UUID 	`gorm:"type:uuid; index" json:"user_id"`
	WalletID 		uuid.UUID 	`gorm:"type:uuid; index" json:"wallet_id"`
	Title 				string 		`json:"title"`
	Amount 				float64 	`json:"amount"`
	Type 				string 		`json:"type"`
	SubType 			string 		`json:"sub_type"`
	Description 		string 		`json:"description"`
	Date 				time.Time 	`json:"date"`
}