package controllers

import (
	"cashflow-backend/app/models"
	"cashflow-backend/pkg/configs"

	"github.com/gofiber/fiber/v2"
)

func GetAllWallets(c *fiber.Ctx) error {
	var wallets []models.Wallet
	results := configs.DB.Preload("Transactions").Limit(20).Find(&wallets)
	if results.Error != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Tidak Menemukan Wallet", "error": results.Error.Error()})
	}
	return c.Status(200).JSON(wallets)
}

func UpdateWallet(c *fiber.Ctx) error {
	id := c.Params("id")
	var wallet models.Wallet
	if result := configs.DB.First(&wallet, "wallet_id = ?", id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Wallet not found", "error": result.Error.Error()})
	}
	if err := c.BodyParser(&wallet); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid Data", "error": err.Error()})
	}
	configs.DB.Save(&wallet)
	return c.Status(200).JSON(wallet)
}

func DeleteWallet(c *fiber.Ctx) error{
	id := c.Params("id")
	var wallet models.Wallet
	if result := configs.DB.First(&wallet, "wallet_id = ?", id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Wallet not found", "error": result.Error.Error()})
	}
	configs.DB.Delete(&wallet)
	return c.Status(200).JSON(fiber.Map{"message": "Wallet deleted successfully"})
}