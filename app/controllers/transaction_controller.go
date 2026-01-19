package controllers

import (
	"cashflow-backend/app/models"
	"cashflow-backend/pkg/configs"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	// "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)


type TransactionInput struct {
    WalletID    string  `json:"wallet_id" validate:"required"` 
    Title       string  `json:"title" validate:"required"`
    Amount      float64 `json:"amount" validate:"required"`
    Type        string  `json:"type" validate:"required,oneof=income expense"`      
    SubType     string  `json:"sub_type" validate:"required"`
    Description string  `json:"description"`
    Date        time.Time  `json:"date"`//validate:"required"` nti tambahin kalo udh ada frontend
}



func CreateTransaction(c *fiber.Ctx) error {
    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Unauthorized / Invalid Token",
        })
    }

    var input TransactionInput
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Format data tidak valid",
            "error":   err.Error(),
        })
    }

    if err:= validate.Struct(input); err != nil{
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Data input tidak lengkap atau salah",
            "error": err.Error(), 
        })
    }

    parsedWalletID, err := uuid.Parse(input.WalletID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Wallet ID tidak valid (bukan format UUID)",
        })
    }

    tx := configs.DB.Begin()

    var wallet models.Wallet
    if err := tx.Where("wallet_id = ? AND user_id = ?", parsedWalletID, userID).First(&wallet).Error; err != nil {
        tx.Rollback() // Batalin
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Wallet tidak ditemukan atau bukan milik Anda",
        })
    }

    if input.Type == "expense" && wallet.Balance < input.Amount {
        tx.Rollback() // Batalin
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Saldo tidak mencukupi untuk melakukan transaksi ini",
            "current_balance": wallet.Balance,
        })
    }

    finalDate:= input.Date

    if finalDate.IsZero(){
        finalDate = time.Now()
    }
    newTransaction := models.Transaction{
        UserID:      userID,         // Dari Token
        WalletID:    parsedWalletID, // Dari hasil parsing UUID di atas
        Title:       input.Title,
        Amount:      input.Amount,
        Type:        input.Type,
        SubType:     input.SubType,
        Description: input.Description,
        Date:        finalDate,
    }

    if err := tx.Create(&newTransaction).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Gagal menyimpan data transaksi",
        })
    }

    if input.Type == "expense" {
        wallet.Balance -= input.Amount
    } else if input.Type == "income" {
        wallet.Balance += input.Amount
    }

    if err := tx.Save(&wallet).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Gagal mengupdate saldo wallet",
        })
    }

    tx.Commit()

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "message":     "Transaksi berhasil dibuat",
        "data":        newTransaction,
        "new_balance": wallet.Balance, 
    })
}

func GetTransactions(c *fiber.Ctx) error {
	var results []WalletResponse
	
	if err := configs.DB.Model(&models.Wallet{}).Find(&results).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data"})
    }
	return c.Status(200).JSON(results)
}

func GetWalletWithTransactions(c *fiber.Ctx) error{
    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Unauthorized / Invalid Token",
        })
    }

    walletId, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Wallet ID tidak valid (bukan format UUID)",
            "wallet_id": c.Params("wallet_id"),
            "error": err.Error(),
        })
    }

    page,_ := strconv.Atoi(c.Query("page", "1"))
    if page <= 0 { 
        page = 1 
    }
    limit, _ := strconv.Atoi(c.Query("limit", "20"))
    if limit <= 0 {
        limit = 20
    }
    if limit > 100 {
        limit = 100 
    }

    offset := (page - 1) * limit


    var wallet models.Wallet

    err = configs.DB.
    Preload("Transactions", func(db *gorm.DB) *gorm.DB{
        return db.Order("created_at DESC").Offset(offset).Limit(limit)
    }).
    Where("wallet_id = ? AND user_id = ?", walletId, userID).First(&wallet).Error; 
    if  err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "message": "Wallet not found", "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Data Wallet berhasil diambil",
        "meta": fiber.Map{
            "page": page,
            "limit": limit,
        },
        "data":    wallet,
    })
}


func GetTransactionByID(c *fiber.Ctx) error {
    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Unauthorized / Invalid Token",
        })
    }

	id := c.Params("id")
	var transaction models.Transaction

	if result := configs.DB.First(&transaction,"transaction_id = ?", id).Where(&userID); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}

	return c.Status(200).JSON(transaction)
}

func UpdateTransaction(c *fiber.Ctx) error{
    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Unauthorized / Invalid Token",
        })
    }

	id := c.Params("id")
	var transaction models.Transaction

	if result := configs.DB.First(&transaction,"transaction_id = ?", id).Where(&userID); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}

	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid Data",
		})
	}

	configs.DB.Save(&transaction)

	return c.Status(200).JSON(transaction)
}

func SoftDeleteTransaction(c *fiber.Ctx) error{
	id := c.Params("id")
	var transaction models.Transaction

	if result := configs.DB.First(&transaction,"transaction_id = ?", id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}
	transaction.DeletedAt.Valid = true
	configs.DB.Save(&transaction)

	return c.Status(200).JSON(fiber.Map{
		"message": "Transaction deleted successfully",
	})
}