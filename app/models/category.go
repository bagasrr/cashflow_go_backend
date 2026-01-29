package models

import (
	"time"

	"github.com/google/uuid"
)
type TransactionType string

const (
    TypeIncome  TransactionType = "income"
    TypeExpense TransactionType = "expense"
)

type Category struct {
    CategoryID uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"category_id"`
    Name       string          `gorm:"type:varchar(100);not null" json:"name"`
    
    // 1. Pembeda Pemasukan/Pengeluaran (Biar pas filter gampang)
    Type       TransactionType `gorm:"type:varchar(20);not null" json:"type"` // "income" atau "expense"
    
    // 2. Self-Referencing buat SUBTYPE
    // Kalau NULL = Induk Kategori (Contoh: Makanan)
    // Kalau ISI = Sub Kategori (Contoh: GoFood, masak ParentID-nya si ID Makanan)
    ParentID   *uuid.UUID      `gorm:"type:uuid;index" json:"parent_id"`
    Parent     *Category       `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
    Children   []Category      `gorm:"foreignKey:ParentID" json:"children,omitempty"`

    // 3. Ownership (Siapa yang punya?)
    // Kalau UserID & GroupID NULL semua -> DEFAULT SYSTEM
    UserID     *uuid.UUID      `gorm:"type:uuid;index" json:"user_id,omitempty"` // Custom User
    GroupID    *uuid.UUID      `gorm:"type:uuid;index" json:"group_id,omitempty"` // Custom Group

    CreatedAt  time.Time       `json:"created_at"`
    UpdatedAt  time.Time       `json:"updated_at"`
}