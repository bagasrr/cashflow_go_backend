package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wallet struct {
	WalletID 	uuid.UUID `gorm:"type:uuid; default:gen_random_uuid(); primaryKey" json:"wallet_id"`

	UserID 		uuid.UUID `gorm:"type:uuid; not null" json:"user_id"`

	GroupID 	uuid.UUID `gorm:"type:uuid; null" json:"group_id"`

	Name 		string `json:"name"`
	Balance 	float64 `gorm:"default:0" json:"balance"`
	Currency 	string `gorm:"type:varchar(10); default:'IDR'" json:"currency"`
	Transactions []Transaction `gorm:"foreignKey:WalletID" json:"transactions,omitempty"`

	CreatedAt 	time.Time      `json:"created_at"`
	UpdatedAt 	time.Time      `json:"updated_at"`
	DeletedAt 	gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}