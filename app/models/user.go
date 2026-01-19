package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 1. Konstanta Role (Hanya untuk Jabatan/Otoritas)
const (
    RoleOwner int8 = 0 // Dewa
    RoleAdmin int8 = 1 // Bisa atur user
    RoleUser  int8 = 2 // Pengguna aplikasi
)

// 2. Konstanta Subscription (Untuk Fitur)
const (
    PlanFree    string = "free"
    PlanPremium string = "premium"
    PlanGold    string = "gold" // Siapa tau nanti ada level di atas premium
)

type User struct {
    UserID    uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"user_id"`
    Username  string         `json:"username"`
    Email     string         `json:"email"`
    Password  string         `json:"password"`
    
    UserRole  int8           `gorm:"default:2;not null" json:"user_role"`
    
    SubscriptionPlan string  `gorm:"default:'free';type:varchar(20)" json:"subscription_plan"`
    
    SubscriptionExp  *time.Time `json:"subscription_exp"`
	Wallets   []Wallet       `gorm:"foreignKey:UserID" json:"wallets,omitempty"`
	Transactions []Transaction `gorm:"foreignKey:UserID" json:"transactions,omitempty"`

    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// Helper: Cek apakah user Premium (Bisa dipakai di Controller nanti)
func (u *User) IsPremium() bool {
    return u.SubscriptionPlan == PlanPremium || u.SubscriptionPlan == PlanGold
}

// Helper: Cek apakah user Admin atau Owner
func (u *User) HasAccess(requiredRole int8) bool {
	return u.UserRole <= requiredRole
}