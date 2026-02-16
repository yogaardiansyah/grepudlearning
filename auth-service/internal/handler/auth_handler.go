// ===== FILE: ./internal/handler/auth_handler.go =====
package handler

import (
	"auth-service/internal/repository" // Pastikan import ini ada
	"auth-service/internal/service"
	"auth-service/internal/utils"      // Pastikan import ini ada
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Struct untuk Login/Register/Verify
type AuthRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// --- TAMBAHKAN STRUCT INI (YANG HILANG) ---
type ReceiptRequest struct {
	UserID   int64   `json:"user_id"`
	OrderID  string  `json:"order_id"`
	Amount   float64 `json:"amount"`
	ItemName string  `json:"item_name"`
}
// ------------------------------------------

func Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}
	if err := service.Register(req.Username, req.Email, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Registrasi berhasil, cek email untuk kode OTP"})
}

func Verify(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}
	if err := service.VerifyEmail(req.Email, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Akun terverifikasi, silakan login"})
}

func Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}
	
	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = "unknown"
	}

	at, rt, err := service.Login(req.Email, req.Password, deviceID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set HttpOnly Cookie (Path: /auth/refresh)
	c.SetCookie("refresh_token", rt, 3600*24*28, "/auth/refresh", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{"access_token": at})
}

func Refresh(c *gin.Context) {
	rt, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		return
	}
	deviceID := c.GetHeader("X-Device-ID")

	newAt, newRt, err := service.RotateRefreshToken(rt, deviceID)
	if err != nil {
		c.SetCookie("refresh_token", "", -1, "/auth/refresh", "localhost", false, true)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
		return
	}

	c.SetCookie("refresh_token", newRt, 3600*24*28, "/auth/refresh", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"access_token": newAt})
}

func Logout(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/auth/refresh", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// Handler Internal: Dipanggil oleh Payment Service
func SendReceipt(c *gin.Context) {
	var req ReceiptRequest // Sekarang Struct ini sudah ada definisinya di atas
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	// 1. Ambil Data User (Email & Username) dari DB Auth
	// Pastikan function GetUserByID sudah ada di repository/auth_repo.go
	user, err := repository.GetUserByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 2. Kirim Email secara Asynchronous (Go Routine) agar tidak blocking
	go func() {
		// Pastikan function SendReceiptEmail sudah ada di utils/email.go
		err := utils.SendReceiptEmail(user.Email, user.Username, req.OrderID, req.Amount, req.ItemName)
		if err != nil {
			fmt.Println("❌ Gagal kirim email receipt:", err)
		} else {
			fmt.Println("✅ Email receipt terkirim ke:", user.Email)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Receipt processed"})
}