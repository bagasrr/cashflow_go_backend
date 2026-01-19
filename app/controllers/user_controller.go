package controllers

import (
	"cashflow-backend/app/models"
	"cashflow-backend/pkg/configs"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

type InputValid struct {
	Username string `json:"username" validate:"required,min=6,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=7"`
}
type UserResponse struct {
	UserID string `json:"user_id"`
	Username string `json:"username"`
	Email string `json:"email"`
	UserRole int8 `json:"user_role"`
	SubscriptionPlan string `json:"subscription_plan"`
	SubscriptionExp *time.Time `json:"subscription_exp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Wallets []WalletResponse `json:"wallet"`
	RoleText string `json:"role_text"`
}

type WalletResponse struct {
        WalletID string  `json:"wallet_id"`
        Name     string  `json:"wallet_name"` 
        Balance  float64 `json:"balance"`
        Currency string  `json:"currency"`

}


func RegisterUser(c *fiber.Ctx) error {

    var input InputValid
    if err := c.BodyParser(&input); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Input tidak valid"})
    }
	
	if err := validate.Struct(input); err != nil {
        // Trik biar pesan errornya enak dibaca (optional)
        return c.Status(400).JSON(fiber.Map{
            "error": "Data tidak lengkap atau format salah",
            "detail": err.Error(), // Ini akan ngasih tau field mana yang kurang
        })
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

    user := models.User{
        Username: input.Username,
        Email:    input.Email,
        Password: string(hashedPassword),
        UserRole: models.RoleUser, // Default 2
        SubscriptionPlan: models.PlanFree,
    }
    tx := configs.DB.Begin()

    if err := tx.Create(&user).Error; err != nil {
        tx.Rollback() // Batalkan semua jika gagal
        return c.Status(500).JSON(fiber.Map{"error": "Gagal mendaftar, email mungkin sudah ada"})
    }

    defaultWallet := models.Wallet{
        UserID:   user.UserID, // Ambil ID dari user yang baru dibuat di atas
        Name:     "My First Wallet", // Nama default
        Balance:  0,
        Currency: "IDR",
    }

    if err := tx.Create(&defaultWallet).Error; err != nil {
        tx.Rollback() // PENTING: Jika wallet gagal, user juga dihapus
        return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat wallet default"})
    }

    tx.Commit()

    // Response
    user.Password = ""
    return c.Status(201).JSON(fiber.Map{
        "message": "Registrasi berhasil & Wallet default dibuat",
        "data":    user,
        "wallet":  defaultWallet, // Opsional: kasih tau wallet apa yang dibuat
    })
}

func GetMyProfile(c *fiber.Ctx) error {
    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Token rusak/ID tidak valid"})
    }
    fmt.Println(">>> DEBUG: ID dari Token adalah:", userID)

    var user models.User
	result := configs.DB.Preload("Wallets").First(&user, "user_id = ?", userID)
    if result.Error != nil {
        // CCTV 2: Lihat error apa yang dibilang database
        fmt.Println(">>> DEBUG: Error Database:", result.Error) 
        return c.Status(404).JSON(fiber.Map{
            "error": "User tidak ditemukan di Database",
            "debug_message": result.Error.Error(),
        })
    }
	var walletRes []WalletResponse
	for _,w := range user.Wallets {
		walletRes = append(walletRes, WalletResponse{
            WalletID: w.WalletID.String(),
			Name: w.Name,
			Balance: w.Balance,
			Currency: w.Currency,
		})
	}

	roleText := "User"
	if user.UserRole == 0 {
		roleText = "Owner"
	}else if user.UserRole == 1 {
        roleText = "Admin"
    }

	response := UserResponse{
		UserID: user.UserID.String(),
		Email: user.Email,
		Username: user.Username,
		UserRole: user.UserRole,
		SubscriptionPlan: user.SubscriptionPlan,
		SubscriptionExp: user.SubscriptionExp,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		RoleText: roleText,
		Wallets: walletRes,
	}
    return c.JSON(response)
}

func UpgradeUserPlan(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Upgrade User Plan"})
}


func GetAllActiveUser(c *fiber.Ctx) error {
    userId, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Invalid token",
        })
    }

    var requestor models.User
    reqval := configs.DB.First(&requestor, "user_id = ?", userId)
    if reqval.Error != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "message": "Unauthorized",
        })
    } 

    if requestor.UserRole > 1 {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "message": "Unauthorized",
        })
    }
   
    var users []models.User
    
    result := configs.DB.Preload("Wallets").Find(&users)
    if result.Error != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "message": "Failed to fetch users",
        })
    }

    var res []UserResponse
    for _, v := range users {
        var walletRes []WalletResponse
        for _, w := range v.Wallets {
            walletRes = append(walletRes, WalletResponse{
                WalletID: w.WalletID.String(),
                Name:     w.Name,
                Balance:  w.Balance,
                Currency: w.Currency,
            })
        }
        res = append(res, UserResponse{
            UserID:           v.UserID.String(),
            Email:            v.Email,
            Username:         v.Username,
            UserRole:         v.UserRole,
            SubscriptionPlan: v.SubscriptionPlan,
            SubscriptionExp:  v.SubscriptionExp,
            CreatedAt:        v.CreatedAt,
            UpdatedAt:        v.UpdatedAt,
            Wallets:          walletRes,
        })
    }

    return c.JSON(fiber.Map{
        "status": "success",
        "data":   res,
    })
}

func DeleteUser(c *fiber.Ctx) error {
    userId := c.Params("id")
    var user models.User
	result := configs.DB.Delete(&user, "user_id = ?", userId)
    if result.Error != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "message": "Failed to delete user",
        })
    }
    return c.JSON(fiber.Map{
        "status": "success",
        "message": "User deleted successfully",
    })
    
}