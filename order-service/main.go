// ===== FILE: ./order-service/main.go =====
package main

import (
	"log"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// Model Sederhana (In-Memory DB)
type Order struct {
	ID     string  `json:"id"`
	UserID string  `json:"user_id"`
	Item   string  `json:"item"`
	Price  float64 `json:"price"`
	Status string  `json:"status"` // pending, paid
}

var (
	orders = make(map[string]Order)
	mu     sync.Mutex
	nextID = 1
)

func main() {
	r := gin.Default()

	// Endpoint: Buat Pesanan
	r.POST("/order/create", func(c *gin.Context) {
		var req struct {
			Item  string  `json:"item"`
			Price float64 `json:"price"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid Input"})
			return
		}

		userID := c.GetHeader("X-User-ID") // Didapat dari Gateway

		mu.Lock()
		id := strconv.Itoa(nextID)
		nextID++
		newOrder := Order{ID: id, UserID: userID, Item: req.Item, Price: req.Price, Status: "pending"}
		orders[id] = newOrder
		mu.Unlock()

		c.JSON(201, newOrder)
	})

	// Endpoint: List Pesanan User
	r.GET("/order/list", func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		var userOrders []Order

		mu.Lock()
		for _, o := range orders {
			if o.UserID == userID {
				userOrders = append(userOrders, o)
			}
		}
		mu.Unlock()
		c.JSON(200, userOrders)
	})

	// Endpoint Internal: Update Status (Dipanggil oleh Payment Service)
	r.POST("/order/internal/update-status", func(c *gin.Context) {
		var req struct {
			OrderID string `json:"order_id"`
			Status  string `json:"status"`
		}
		c.ShouldBindJSON(&req)

		mu.Lock()
		if val, ok := orders[req.OrderID]; ok {
			val.Status = req.Status
			orders[req.OrderID] = val
			mu.Unlock()
			c.JSON(200, gin.H{"message": "updated"})
		} else {
			mu.Unlock()
			c.JSON(404, gin.H{"error": "not found"})
		}
	})

	log.Println("📦 Order Service running on port 8081")
	r.Run(":8081")
}