// ===== FILE: ./payment-service/main.go =====
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv" // Tambahkan ini

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/payment/pay", func(c *gin.Context) {
		// Request dari Frontend
		var req struct {
			OrderID  string  `json:"order_id"`
			Amount   float64 `json:"amount"`
		}
		
		// Header User ID dari Gateway (String)
		userIDStr := c.GetHeader("X-User-ID")
		
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid Input"})
			return
		}

		log.Printf("💰 Processing payment for Order %s Amount %.2f", req.OrderID, req.Amount)

		// 1. Update Status di Order Service
		updatePayload := map[string]string{
			"order_id": req.OrderID,
			"status":   "paid",
		}
		jsonBody, _ := json.Marshal(updatePayload)
		resp, err := http.Post("http://localhost:8081/order/internal/update-status", "application/json", bytes.NewBuffer(jsonBody))
		
		if err != nil || resp.StatusCode != 200 {
			c.JSON(500, gin.H{"error": "Payment success but failed to update order"})
			return
		}

		// 2. TRIGGER KIRIM EMAIL KE AUTH SERVICE
		// Kita pakai Goroutine agar user tidak perlu menunggu email terkirim
		go func(oID string, amt float64, uIDStr string) {
			// Convert UserID string ke int64
			uID, _ := strconv.ParseInt(uIDStr, 10, 64)

			receiptPayload := map[string]interface{}{
				"user_id":   uID,
				"order_id":  oID,
				"amount":    amt,
				"item_name": "Makanan Lezat", // Idealnya ambil detail item dari Order Service dulu, tapi kita hardcode dulu untuk demo
			}
			receiptJson, _ := json.Marshal(receiptPayload)

			// Tembak Auth Service
			_, err := http.Post("http://localhost:8080/auth/internal/send-receipt", "application/json", bytes.NewBuffer(receiptJson))
			if err != nil {
				log.Println("⚠️ Gagal trigger receipt:", err)
			} else {
				log.Println("📨 Request kirim struk dikirim ke Auth Service")
			}
		}(req.OrderID, req.Amount, userIDStr)

		c.JSON(200, gin.H{"message": "Payment Successful", "order_id": req.OrderID})
	})

	log.Println("💳 Payment Service running on port 8082")
	r.Run(":8082")
}