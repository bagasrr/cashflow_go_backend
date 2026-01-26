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
    Username string `json:"username" validate:"required,min=5,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=5"`
}

type UserResponse struct {
    UserID           string           `json:"user_id"`
    Username         string           `json:"username"`
    Email            string           `json:"email"`
    UserRole         models.UserRole  `json:"user_role"` 
    RoleText         string           `json:"role_text"`
    SubscriptionPlan string           `json:"subscription_plan"`
    SubscriptionExp  *time.Time       `json:"subscription_exp"`
    CreatedAt        time.Time        `json:"created_at"`
    UpdatedAt        time.Time        `json:"updated_at"`
    Wallets          []WalletResponse `json:"wallet"`
}
type UpgradePlanInput struct {
    Plan string `json:"plan" validate:"required,oneof=premium gold"`
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
    
    // ... (Validasi Duplikat & Validator Struct tetap sama) ...
    var existingUser models.User
    if err := configs.DB.Where("email = ? OR username = ?", input.Email, input.Username).First(&existingUser).Error; err == nil {
        return c.Status(409).JSON(fiber.Map{"error": "Email atau Username sudah terdaftar"})
    }
    if err := validate.Struct(input); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Data tidak lengkap", "detail": err.Error()})
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

    // FIX 1: Siapkan variabel role default (UserDefault = 3)
    // Jangan langsung assign angka, harus via variabel biar bisa diambil alamat memorinya (&)
    defaultRole := models.UserDefault

    user := models.User{
        Username: input.Username,
        Email:    input.Email,
        Password: string(hashedPassword),
        UserRole: &defaultRole, 
        SubscriptionPlan: models.PlanFree,
    }

    tx := configs.DB.Begin()
    if err := tx.Create(&user).Error; err != nil {
        tx.Rollback()
        fmt.Println(err)
        return c.Status(500).JSON(fiber.Map{"message": "Gagal mendaftar", "error" : err.Error()})
    }

    defaultWallet := models.Wallet{
        UserID:   user.UserID,
        Name:     "My Wallet",
        Balance:  0,
        Currency: "IDR",
    }

    if err := tx.Create(&defaultWallet).Error; err != nil {
        tx.Rollback()
        return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat wallet default"})
    }

    tx.Commit()

    user.Password = ""
    return c.Status(201).JSON(fiber.Map{
        "message": "Registrasi berhasil",
        "data":    user,
    })
}

func GetMyProfile(c *fiber.Ctx) error {
    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Token rusak"})
    }

    var user models.User
    result := configs.DB.Preload("Wallets").First(&user, "user_id = ?", userID)
    if result.Error != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User tidak ditemukan"})
    }

    // Mapping Wallet Response
    var walletRes []WalletResponse
    for _, w := range user.Wallets {
        walletRes = append(walletRes, WalletResponse{
            WalletID: w.WalletID.String(),
            Name:     w.Name,
            Balance:  w.Balance,
            Currency: w.Currency,
        })
    }

    // FIX 2: Logic Role Text & Pointer Handling
    // roleText := "User"
    // roleValue := models.UserDefault // Default kalau nil

    // if user.UserRole != nil {
    //     roleValue = *user.UserRole // Ambil value dari pointer
        
    //     // Sesuaikan dengan konstanta baru (1=Owner, 2=Admin)
    //     if roleValue == models.UserOwner {
    //         roleText = "Owner"
    //     } else if roleValue == models.UserAdmin {
    //         roleText = "Admin"
    //     }
    // }

    response := UserResponse{
        UserID:           user.UserID.String(),
        Email:            user.Email,
        Username:         user.Username,
        UserRole:         *user.UserRole, // <-- Masukkan value integer (bukan pointer)
        SubscriptionPlan: user.SubscriptionPlan,
        SubscriptionExp:  user.SubscriptionExp,
        CreatedAt:        user.CreatedAt,
        UpdatedAt:        user.UpdatedAt,
        RoleText:         user.UserRole.String(),
        Wallets:          walletRes,
    }
    return c.JSON(response)
}

func GetAllActiveUser(c *fiber.Ctx) error {
    userId, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid token"})
    }

    var requestor models.User
    if err := configs.DB.First(&requestor, "user_id = ?", userId).Error; err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
    }

    // FIX 3: Validasi Role (Admin & Owner boleh akses)
    // Cek nil dulu, baru cek value
    // Logic: Kalau role > Admin (berarti User Biasa/3), tolak.
    if requestor.UserRole == nil || *requestor.UserRole > models.UserAdmin {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "message": "Unauthorized: Only Admin/Owner allowed",
        })
    }
   
    var users []models.User
    if err := configs.DB.Preload("Wallets").Find(&users).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch users"})
    }

    var res []UserResponse
    for _, v := range users {
        
        // FIX 4: Handle pointer role di dalam loop
        rVal := models.UserDefault
        rTxt := "User"
        if v.UserRole != nil {
            rVal = *v.UserRole
            if rVal == models.UserOwner { rTxt = "Owner" }
            if rVal == models.UserAdmin { rTxt = "Admin" }
        }

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
            UserRole:         rVal, // <-- Value
            RoleText:         rTxt, // <-- Text
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

// ... DeleteUser sama saja, aman ...
func DeleteUser(c *fiber.Ctx) error {
    userId := c.Params("id")
    var user models.User

    // 1. Cek dulu usernya ada atau nggak (Opsional tapi bagus buat UX)
    if err := configs.DB.First(&user, "user_id = ?", userId).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }

    // 2. Lakukan HARD DELETE dengan Cascade
    // .Unscoped() = Abaikan deleted_at, hapus barisnya dari fisik tabel
    // Karena di Model sudah ada 'OnDelete:CASCADE', database akan otomatis menghapus wallet & transaksi terkait.
    result := configs.DB.Unscoped().Delete(&user)

    if result.Error != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Gagal menghapus user", 
            "detail": result.Error.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "status": "success",
        "message": "User dan seluruh datanya berhasil dihapus permanen",
    })
}
func UpgradeUserPlan(c *fiber.Ctx) error {
    // 1. Ambil User ID dari Token
    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
    }

    // 2. Parse Input Body
    var input UpgradePlanInput
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format input salah"})
    }

    // 3. Validasi Input (Cuma boleh 'premium' atau 'gold')
    if err := validate.Struct(input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Plan tidak valid (pilih: premium, gold)",
            "detail": err.Error(),
        })
    }

    // 4. Cari User di Database
    var user models.User
    if err := configs.DB.First(&user, "user_id = ?", userID).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User tidak ditemukan"})
    }

    // 5. Logic Durasi Berlangganan
    // Misal: Premium = 30 Hari, Gold = 1 Tahun (Contoh saja)
    var duration time.Duration
    if input.Plan == models.PlanGold {
        duration = 365 * 24 * time.Hour // 1 Tahun
    } else {
        duration = 30 * 24 * time.Hour // 30 Hari (Premium)
    }

    // Hitung tanggal kedaluwarsa baru
    // Kalau user masih aktif, tambah dari tanggal exp terakhir. Kalau sudah mati, dari NOW.
    newExpDate := time.Now().Add(duration)
    if user.SubscriptionExp != nil && user.SubscriptionExp.After(time.Now()) {
        // Jika masih aktif, perpanjang dari tanggal exp yang ada
        newExpDate = user.SubscriptionExp.Add(duration)
    }

    // 6. Update Data User
    user.SubscriptionPlan = input.Plan
    user.SubscriptionExp = &newExpDate // Assign pointer

    // Simpan ke Database
    if err := configs.DB.Save(&user).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengupdate plan"})
    }

    return c.JSON(fiber.Map{
        "message": "Berhasil upgrade plan",
        "data": fiber.Map{
            "user_id": user.UserID,
            "plan": user.SubscriptionPlan,
            "expires_at": user.SubscriptionExp,
        },
    })
}