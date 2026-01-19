package controllers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func getUserIDFromToken(c *fiber.Ctx) (uuid.UUID, error) {
    // 1. Cek apakah token ada di locals (Defensive Programming)
    userLocals := c.Locals("user")
    if userLocals == nil {
        return uuid.Nil, fmt.Errorf("token tidak ditemukan di context")
    }

    // 2. Casting ke *jwt.Token
    userToken, ok := userLocals.(*jwt.Token)
    if !ok {
        return uuid.Nil, fmt.Errorf("format token salah")
    }

    // 3. Ambil Claims
    claims, ok := userToken.Claims.(jwt.MapClaims)
    if !ok {
        return uuid.Nil, fmt.Errorf("gagal membaca claims")
    }

    // 4. Ambil string user_id dengan cara AMAN (Anti-Panic)
    // variable 'ok' akan false jika key tidak ada ATAU bukan string
    idString, ok := claims["user_id"].(string)
    if !ok {
        return uuid.Nil, fmt.Errorf("user_id tidak ditemukan di token atau bukan string")
    }

    // 5. Parse String menjadi UUID
    // Di sini validasi terakhir. Kalau stringnya "aku-ganteng", 
    // uuid.Parse akan return error karena itu bukan format UUID.
    userID, err := uuid.Parse(idString)
    if err != nil {
        return uuid.Nil, fmt.Errorf("format user_id di token bukan UUID valid")
    }

    return userID, nil
}