package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 1. Enum Role (Sudah Benar)
type GroupRole int8

const (
    GroupOwner  GroupRole = 1
    GroupAdmin  GroupRole = 2
    GroupParticipan GroupRole = 3
    GroupViewer GroupRole = 4
)

// Helper String (Sudah Benar)
func (r GroupRole) String() string {
    switch r {
    case GroupOwner:
        return "Group Owner"
    case GroupAdmin:
        return "Group Admin"
    case GroupParticipan:
        return "Group Member"
    case GroupViewer:
        return "Group Viewer"
    default:
        return "Group Unknown"
    }
}


type Group struct {
    GroupID   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"group_id"`
    Name      string    `gorm:"type:varchar(100);not null" json:"name"`
    Type      string    `gorm:"type:varchar(50);not null" json:"type"`
    CreatedBy uuid.UUID `gorm:"type:uuid;index" json:"created_by"`

    // PERBAIKAN: Ubah []GroupMembers jadi []GroupMember (Singular)
    Members   []GroupMember `gorm:"foreignKey:GroupID" json:"members,omitempty"`
    Wallets   []Wallet      `gorm:"foreignKey:GroupID" json:"wallets,omitempty"`

    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// PERBAIKAN: Ubah nama struct jadi GroupMember (Singular)
type GroupMember struct {
    MemberID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"member_id"`

    GroupID  uuid.UUID `gorm:"type:uuid;index:idx_group_user,unique" json:"group_id"`
    UserID   uuid.UUID `gorm:"type:uuid;index:idx_group_user,unique" json:"user_id"`
    
    Role     GroupRole `gorm:"type:smallint;default:2" json:"role"`

    // Relasi ke User
    User     User      `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE" json:"user"`

    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}