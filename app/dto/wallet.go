package dto

type WalletResponse struct {
    WalletID string  `json:"wallet_id" example:"a0b1c2d3-e4f5-g6h7-i8j9-k0l1m2n3"`
    Name     string  `json:"wallet_name" example:"wallet1"`
    Balance  float64 `json:"balance" example:"100000"`
    Currency string  `json:"currency" example:"IDR"`
}