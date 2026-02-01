package dto

import "time"

type TransactionInput struct {
    WalletID    string  `json:"wallet_id" validate:"required"` 
    Title       string  `json:"title" validate:"required"`
    Amount      float64 `json:"amount" validate:"required"`
    CategoryID  string    `json:"category_id" validate:"required,uuid"` 
    Description string  `json:"description"`
    Date        time.Time  `json:"date"`//validate:"required"` nti tambahin kalo udh ada frontend
}
