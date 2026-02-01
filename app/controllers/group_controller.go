package controllers

import (
	"cashflow-backend/app/dto"
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
    User     dto.UserResponse `json:"user"`
}
type GroupResponse struct {
    GroupID string                `json:"group_id"`
    Name    string                `json:"name"`
    Type    string                `json:"type"`
    Members []MemberResponse        `json:"members"`
}

type AddPeopleInput struct {
    GroupID string `json:"group_id" validate:"required,uuid"`
    Email   string `json:"email" validate:"required,email"`
}

// CreateGroup godoc
// @Summary      Membuat Group Baru
// @Description  Endpoint untuk User membuat grup baru dan otomatis menjadi group-owner.
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body controllers.GroupInput true "Data Group (Minimal 1 member)"
// @Success      201  {object} controllers.GroupResponse
// @Failure      400  {object} map[string]interface{} "Error validasi / ID member salah"
// @Failure      401  {object} map[string]interface{} "Token tidak valid"
// @Failure      500  {object} map[string]interface{} "Server error"
// @Router       /groups [post]
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
                Role:     m.Role.String(), // 0 jadi "Owner", 2 jadi "Member"
                User: dto.UserResponse{
                    UserID:   m.User.UserID.String(),
                    Username: m.User.Username,
                    Email:    m.User.Email,
                    UserRole: *m.User.UserRole,
                    RoleText: m.User.UserRole.String(),
                    SubscriptionPlan: m.User.SubscriptionPlan,
                    SubscriptionExp: m.User.SubscriptionExp,
                    CreatedAt : m.User.CreatedAt,
                    UpdatedAt: m.User.UpdatedAt,
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


// Add People to Group By email
// @Summary      Menambahkan anggota baru ke Group 
// @Description  Menambahkan anggota baru ke Group berdasarkan email
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Param        request body AddPeopleInput true "AddPeopleInput"
// @Security     BearerAuth
// @Success      200  {object} map[string]interface{} "Berhasil menambahkan anggota"    
// @Failure      400  {object} map[string]interface{} "Input Tidak Valid"
// @Failure      401  {object} map[string]interface{} "Unauthorized / Invalid Token"
// @Failure      403  {object} map[string]interface{} "Anda bukan anggota grup ini"
// @Failure      404  {object} map[string]interface{} "Email user tidak ditemukan di sistem."
// @Failure      409  {object} map[string]interface{} "User ini sudah menjadi anggota grup."
// @Failure      500  {object} map[string]interface{} "Gagal menambahkan anggota baru."
// @Router       /groups/add-new-member [post]
func AddPeopleToGroupByEmail(c *fiber.Ctx) error {
    reqID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Unauthorized / Invalid Token",
        })
    }
    var input AddPeopleInput
    if err := c.BodyParser(&input).Error; err != nil{
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message" : "Input Tidak Valid",
        })
    }

    tx := configs.DB.Begin()
    var isGroupMember models.GroupMember
    if err := tx.Where("group_id = ? , user_id = ?", input.GroupID, reqID).First(&isGroupMember).Error; err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message" : "Anda bukan anggota grup ini",
        })
    }

    if isGroupMember.Role > models.GroupAdmin {
        tx.Rollback()
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "message" : "Akses Ditolak ! Hanya Admin atau Owner yang bisa menambahkan anggota.",
        })
    }

    var targetUser models.User
    if err := tx.Where("email = ? ", input.Email).First(&targetUser).Error ; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "message" : "Email user tidak ditemukan di sistem",
        })
    }

    var existingMember models.GroupMember
    check := tx.Where("group_id = ? , user_id = ?").First(&existingMember)
    if check.RowsAffected > 0 {
        tx.Rollback()
        return c.Status(fiber.StatusConflict).JSON(fiber.Map{
            "message" : "User ini sudah menjadi anggota grup",
        })
    }

    newMember := models.GroupMember {
        GroupID : isGroupMember.GroupID,
        UserID : targetUser.UserID,
        Role : models.GroupParticipan,
    }

    if err := tx.Create(&newMember).Error ; err!= nil {
        tx.Rollback()
        c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "message" : "Gagal menambahkan anggota baru",
        })
    }

    tx.Commit()

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message" : "Berhasil menambahkan anggota",
        "data" : fiber.Map{
            "email" : targetUser.Email,
            "username" : targetUser.Username,
            "role" : newMember.Role,
        },
    })

}