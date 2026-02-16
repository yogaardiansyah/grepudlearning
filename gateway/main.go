// ===== FILE: ./gateway/main.go =====
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// Helper: Proxy Request
func proxyRequest(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		remote, _ := url.Parse(target)
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
            
			// UBAH BARIS INI:
			// Gunakan c.Request.URL.Path agar "/order/create" tetap utuh
			req.URL.Path = c.Request.URL.Path 
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
// Middleware: Validasi Token & Inject User ID
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			// Kirim User ID ke Microservice via Header
			c.Request.Header.Set("X-User-ID", fmt.Sprintf("%v", claims["sub"]))
		}
		c.Next()
	}
}

func main() {
	godotenv.Load() // Pastikan ada JWT_SECRET di .env

	r := gin.Default()

	// Setup CORS untuk Gateway
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Device-ID")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 1. Route ke Auth Service (Tanpa Middleware Auth)
	// Request ke /auth/login akan diteruskan ke localhost:8080/auth/login
	r.Any("/auth/*proxyPath", proxyRequest("http://localhost:8080"))

	// 2. Route ke Order Service (Butuh Login)
	r.Any("/order/*proxyPath", AuthMiddleware(), proxyRequest("http://localhost:8081"))

	// 3. Route ke Payment Service (Butuh Login)
	r.Any("/payment/*proxyPath", AuthMiddleware(), proxyRequest("http://localhost:8082"))

	log.Println("🚪 API Gateway running on port 8000")
	r.Run(":8000")
}