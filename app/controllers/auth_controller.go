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

type LoginInput struct {
	Email string `json:"email" example:"username@example.com" validate:"required,email"`
	Password string `json:"password" example:"password" validate:"required,min=5"`	
}
type LoginSuccess struct{
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Njk4MTE0NTIsInJvbGUiOjMsInVzZXJfaWQiOiJkN"`
	Message string `json:"message" example:"Login Success"`
}
type LoginError struct{
	Message string `json:"message" example:"Email atau Password Salah"`
	Error string	`json:"error,omitempty" example:"Record Not Found"`
}
type ServerError struct{
	Message string `json:"message" example:"Server Error"`
}
// Login
// @Summary Login
// @Description Endpoint untuk login
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginInput true "Login Input"
// @Success 200 {object} controllers.LoginSuccess
// @Failure 400 {object} controllers.LoginError
// @Failure 401 {object} controllers.LoginError
// @Failure 500 {object} controllers.ServerError
// @Router /auth/login [post]
func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(LoginError{
			Message: "Input Tidak Valid",
			Error : err.Error(),
	})
	}

	var user models.User
	if err := configs.DB.Where("email = ?", input.Email).First(&user).Error; err!= nil {
		return c.Status(401).JSON(LoginError{
			Message: "Email Atau Password Salah",
			Error : err.Error(),
		})
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return c.Status(401).JSON(LoginError{
			Message: "Email Atau Password Salah",
			Error : err.Error(),
		})
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
	return c.JSON(LoginSuccess{
		Token: t,
		Message: "Login Success",
    })
}