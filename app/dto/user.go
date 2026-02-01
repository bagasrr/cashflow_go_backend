package dto

import (
	"cashflow-backend/app/models"
	"time"
)

// Input Struct
type UpgradePlanInput struct {
    Plan string `json:"plan" validate:"required,oneof=premium gold"`
}

type InputRegisterValid struct {
    Username string `json:"username" validate:"required,min=5,alphanum" example:"johndoe"`
    Email    string `json:"email" validate:"required,email" example:"johndoe@example.com"`
    Password string `json:"password" validate:"required,min=5" example:"john123"`
}
// Resnponse Struct

type UserResponse struct {
    UserID           string           `json:"user_id" example:"a0b1c2d3-e4f5-g6h7-i8j9-k0l1m2n3"`
    Username         string           `json:"username" example:"JohnDoe"`
    Email            string           `json:"email" example:"johndoe@example.com"`
    UserRole         models.UserRole  `json:"user_role" example:"3"` 
    RoleText         string           `json:"role_text" example:"User"`
    SubscriptionPlan string           `json:"subscription_plan" example:"free"`
    SubscriptionExp  *time.Time       `json:"subscription_exp" example:"2023-06-30T00:00:00Z"`
    CreatedAt        time.Time        `json:"created_at" example:"2023-06-30T00:00:00Z"`
    UpdatedAt        time.Time        `json:"updated_at" example:"2023-06-30T00:00:00Z"`
    Wallets          []WalletResponse `json:"wallet"`
}

type RegisterFailed struct {
    Message string `json:"message" example:"Email atau Username sudah terdaftar"`
}
type ResponseSuccess struct {
    Message string `json:"message" example:"Register Berhasil"`
    Data    UserResponse `json:"data"`
}
type ResGetUserSuccess struct {
    Message string `json:"message" example:"Berhasil"`
    Data    []UserResponse `json:"data"`
}
type DefaultMessageResponse struct{
	Message string `json:"message" example:"Token Invalid"`
}

type FailureMessage struct{
    Message string `json:"message" example:"Token Invalid"`
    Error string `json:"error,omitempty" example:"Record Not Found"`
}