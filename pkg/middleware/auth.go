package middleware

import (
	"fmt"
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func Protected() fiber.Handler{
	return jwtware.New(jwtware.Config{
        // Ambil secret dari .env
        SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
        
        // Kalau error (gak ada token / token salah), jalankan fungsi ini:
        ErrorHandler: func(c *fiber.Ctx, err error) error {
			fmt.Println(">>> JWT Error:", err.Error())
            return c.Status(401).JSON(fiber.Map{
                "message": "Unauthorized. Silakan login terlebih dahulu.",
				"error" : err.Error(),
            })
        },
    })
}