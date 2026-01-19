package controllers

import (
	"cashflow-backend/app/models"
	"cashflow-backend/pkg/configs"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input Tidak Valid"})
	}

	var user models.User
	if err := configs.DB.Where("email = ?", input.Email).First(&user).Error; err!= nil {
		return c.Status(401).JSON(fiber.Map{"message" : "email atau password salah"})
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"message" : "email atau password salah"})
	}

	// generate Jwt Tokern
	claims := jwt.MapClaims{
		"user_id" : user.UserID,
		"role" : user.UserRole,
		"exp" : time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
    t, err := token.SignedString([]byte(secret))
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Gagal login"})
    }

    // 5. Kirim Token ke User
    return c.JSON(fiber.Map{
        "message": "Login sukses",
        "token":   t, // Ini "karcis" yang harus disimpan user
    })
}