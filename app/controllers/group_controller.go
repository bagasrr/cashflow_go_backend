package controllers

import (
	"cashflow-backend/app/models"
	"cashflow-backend/pkg/configs"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)


type GroupInput struct {
	Name string `json:"name" validate:"required"`
	GroupType string `json:"group_type" validate:"required"`
	Members []string `json:"members" validate:"required,min=1" `
}
type MemberResponse struct{
	MemberID string   `json:"member_id"`
    Role     string   `json:"role"` // Kita pake string (Owner/Member) bukan angka
    User     UserResponse `json:"user"`
}
type GroupResponse struct {
    GroupID string                `json:"group_id"`
    Name    string                `json:"name"`
    Type    string                `json:"type"`
    Members []MemberResponse `json:"members"`
}
func CreateGroup(c *fiber.Ctx) error {
    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Unauthorized / Invalid Token",
        })
    }

    var input GroupInput
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Format data tidak valid",
            "error":   err.Error(),
        })
    }

    // Validator 'min=1' di struct akan otomatis menolak jika 'members' kosong
    if err := validate.Struct(input); err != nil {
		if input.Members == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Untuk buat group minimal harus ada 1 member.",
			})
		}
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Data input salah. Minimal harus ajak 1 member lain.",
            "error":   err.Error(),
        })
    }

    // Mulai Transaksi
    tx := configs.DB.Begin()

    // 1. Buat Group Header
    group := models.Group{
        Name:      input.Name,
        Type:      input.GroupType,
        CreatedBy: userID,
    }

    if err := tx.Create(&group).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "message": "Gagal membuat group",
            "error":   err.Error(),
        })
    }

    // 2. Masukkan Diri Sendiri sebagai OWNER
    ownerMember := models.GroupMember{
        GroupID: group.GroupID,
        UserID:  userID,
        Role:    models.GroupOwner, // 0: Owner
    }

    if err := tx.Create(&ownerMember).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "message": "Gagal menambahkan owner",
            "error":   err.Error(),
        })
    }

    // 3. Masukkan Member Lain (Undangan)
    // Kita looping array input.Members
    for _, memberIDStr := range input.Members {
        // Parse String ke UUID
        memberUUID, err := uuid.Parse(memberIDStr)
        if err != nil {
            tx.Rollback() // Batalin semua kalau ada ID yang ngaco formatnya
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "message": "Salah satu ID member tidak valid (bukan UUID)",
                "invalid_id": memberIDStr,
            })
        }

        // Cek biar gak invite diri sendiri (double)
        if memberUUID == userID {
            continue 
        }

        newMember := models.GroupMember{
            GroupID: group.GroupID,
            UserID:  memberUUID,
            Role:    models.GroupParticipan, // 2: Member Biasa
        }

        // Simpan ke DB
        // Kalau ID usernya gak ada di tabel Users, ini bakal error Foreign Key
        if err := tx.Create(&newMember).Error; err != nil {
            tx.Rollback() // Batalin Group-nya juga
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "message": "Gagal menambahkan member. Pastikan User ID terdaftar.",
                "detail": err.Error(),
                "failed_user_id": memberIDStr,
            })
        }
    }

    // 4. Commit (Simpan Permanen)
    tx.Commit()

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "message": "Group berhasil dibuat dengan member awal",
        "data": fiber.Map{
            "group_id": group.GroupID,
            "name":     group.Name,
            "total_initial_members": len(input.Members) + 1, // +1 itu Owner
        },
    })
}

func GetAllGroups(c *fiber.Ctx) error {
    
    // _, err := getUserIDFromToken(c) 
    // if err != nil {
    //      return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
    // }

    var groups []models.Group
    
    // 2. Query DB dengan Nested Preload
    // Ambil Group dimana usernya terdaftar sebagai member
    // Logic join: Join group_members -> filter user_id -> load groups
    // Tapi untuk simplenya (kalau kamu mau ambil SEMUA group yg ada di DB):
    if err := configs.DB.
        Preload("Members.User"). // Load user detail di dalam member
        Find(&groups).Error; err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data"})
    }

    // 3. Mapping Data (Manual) dari Model ke Response Struct
    var response []GroupResponse

    for _, g := range groups {
        // Siapkan wadah untuk members di group ini
        var membersRes []MemberResponse
        
        for _, m := range g.Members {
            membersRes = append(membersRes, MemberResponse{
                MemberID: m.MemberID.String(),
                // DISINI KITA PAKE FUNGSI String() TADI
                Role:     m.Role.String(), // 0 jadi "Owner", 2 jadi "Member"
                User: UserResponse{
                    UserID:   m.User.UserID.String(),
                    Username: m.User.Username,
                    Email:    m.User.Email,
                    // Password gak kita masukin, AMAN!
                },
            })
        }

        // Masukkan ke array utama
        response = append(response, GroupResponse{
            GroupID: g.GroupID.String(),
            Name:    g.Name,
            Type:    g.Type,
            Members: membersRes,
        })
    }

    return c.JSON(fiber.Map{
        "message": "Berhasil mendapatkan daftar group",
        "data":    response,
    })
}