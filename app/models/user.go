package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole int8
// 1. Konstanta Role (Hanya untuk Jabatan/Otoritas)
const (
    UserOwner UserRole = 1 // Dewa
    UserAdmin UserRole = 2 // Bisa atur user
    UserDefault UserRole = 3 // Pengguna aplikasi
)
func (u *UserRole) String() string {
    switch *u {
    case UserOwner:
        return "Owner"
    case UserAdmin:
        return "Admin"
    default:
        return "User"
    }
}
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
    
    UserRole  *UserRole           `gorm:"default:3;not null" json:"user_role"`
    
    SubscriptionPlan string  `gorm:"default:'free';type:varchar(20)" json:"subscription_plan"`
    
    SubscriptionExp  *time.Time `json:"subscription_exp"`
    Wallets      []Wallet      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"wallets,omitempty"`
    Transactions []Transaction `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"transactions,omitempty"`
   
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// Helper: Cek apakah user Premium (Bisa dipakai di Controller nanti)
func (u *User) IsPremium() bool {
    return u.SubscriptionPlan == PlanPremium || u.SubscriptionPlan == PlanGold
}

// Helper: Cek apakah user Admin atau Owner
func (u *User) HasAccess(requiredRole UserRole) bool {
    if u.UserRole == nil {
        return false
    }
	return *u.UserRole <= requiredRole
}