package controllers

import (
	"cashflow-backend/app/dto"
	"cashflow-backend/app/models"
	"cashflow-backend/pkg/configs"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	// "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)





func CreateTransaction(c *fiber.Ctx) error {
    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
    }

    var input dto.TransactionInput
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data salah"})
    }

    if err := validate.Struct(input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Parse UUID
    walletUUID, _ := uuid.Parse(input.WalletID)
    categoryUUID, err := uuid.Parse(input.CategoryID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Category ID tidak valid"})
    }

    tx := configs.DB.Begin()

    // 1. Cek Wallet (Pastikan milik User)
    // Note: Nanti logic ini perlu diupdate kalau mau support Group Wallet
    var wallet models.Wallet
    if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("wallet_id = ? AND user_id = ?", walletUUID, userID).
    First(&wallet).Error; err != nil {
        tx.Rollback()

        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Wallet tidak ditemukan"})
    }

    // 2. CEK KATEGORI (STEP BARU & KRUSIAL)
    // Kita butuh tau apakah kategori ini 'expense' atau 'income'
    var category models.Category
    if err := tx.First(&category, categoryUUID).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Kategori tidak ditemukan"})
    }

    // 3. Logic Saldo Berdasarkan Kategori
    // "category.Type" diambil dari Database, bukan input user. Lebih aman.
    if category.Type == "expense" {
        if wallet.Balance < input.Amount {
            tx.Rollback()
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "Saldo tidak cukup",
                "current_balance": wallet.Balance,
            })
        }
        wallet.Balance -= input.Amount
    } else {
        wallet.Balance += input.Amount
    }

    // 4. Buat Object Transaction
    finalDate := input.Date
    if finalDate.IsZero() {
        finalDate = time.Now()
    }

    newTransaction := models.Transaction{
        UserID:      userID,
        WalletID:    walletUUID,
        CategoryID:  categoryUUID, // Simpan ID-nya saja
        Amount:      input.Amount,
        Description: input.Description, // Optional
        Date:        finalDate,
    }

    // 5. Simpan Transaksi
    if err := tx.Create(&newTransaction).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal simpan transaksi"})
    }

    // 6. Update Wallet Balance
    if err := tx.Save(&wallet).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update saldo"})
    }

    tx.Commit()

    // Response Data (Biar Frontend seneng, kita balikin type-nya juga)
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "message": "Berhasil",
        "data": fiber.Map{
            "transaction_id": newTransaction.TransactionID,
            "amount":         newTransaction.Amount,
            "category_name":  category.Name, 
            "type":           category.Type, 
            "new_balance":    wallet.Balance,
        },
    })
}

func GetTransactions(c *fiber.Ctx) error {
	var results []dto.WalletResponse
	
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