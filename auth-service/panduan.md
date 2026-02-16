
# Tutorial Goleng - grepud
---
Kode ini menggunakan:
1.  **Golang + Gin** (Framework).
2.  **PostgreSQL** (Database Utama).
3.  **Redis** (Simpan OTP & Cache).
4.  **Argon2** (Hashing Password Terkuat).
5.  **Universal SMTP** (Gmail/Mailtrap).
6.  **Clean Architecture**.

---

### TAHAP 1: Setup Database (Lakukan di Terminal/CMD)

Kita siapkan "wadah" datanya dulu. Pastikan Postgres & Redis sudah nyala.

1.  Buka terminal, masuk ke Postgres CLI:
    ```bash
    psql -U postgres
    ```
2.  Copy-paste seluruh blok SQL ini ke dalam terminal Postgres:

```sql
-- 1. Buat Database
CREATE DATABASE auth_db;

-- 2. Masuk ke Database
\c auth_db

-- 3. Tabel Users (Support Verifikasi Email)
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Tabel Refresh Tokens (Untuk Rotation & Security)
CREATE TABLE refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    device_id VARCHAR(128) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    absolute_expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Indexing biar cepat
CREATE INDEX idx_refresh_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_user_device ON refresh_tokens(user_id, device_id);

-- 6. Keluar
\q
```

---

### TAHAP 2: Setup Project & Struktur Folder

Buka terminal di lokasi di mana Anda ingin menyimpan project (misal: `Documents/Projects`), lalu jalankan perintah ini satu per satu:

```bash
mkdir auth-service
cd auth-service
go mod init auth-service

# Buat Struktur Folder Otomatis
mkdir -p cmd/api
mkdir -p internal/database
mkdir -p internal/handler
mkdir -p internal/middleware
mkdir -p internal/models
mkdir -p internal/repository
mkdir -p internal/service
mkdir -p internal/utils
```

---

### TAHAP 3: Copy-Paste Kode (Full)

Sekarang buat file sesuai nama yang tertera di header setiap blok kode di bawah ini.

#### 1. File: `internal/models/models.go`
*Definisi bentuk data User dan Token.*

```go
package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
}

type RefreshToken struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"user_id"`
	TokenHash         string     `json:"-"`
	DeviceID          string     `json:"device_id"`
	ExpiresAt         time.Time  `json:"expires_at"`
	AbsoluteExpiresAt time.Time  `json:"absolute_expires_at"`
	RevokedAt         *time.Time `json:"revoked_at"`
}
```

#### 2. File: `internal/database/db.go`
*Koneksi ke Postgres dan Redis.*

```go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var DB *sql.DB
var RDB *redis.Client

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("❌ Failed to open DB:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("❌ DB not reachable:", err)
	}
	log.Println("✅ Connected to PostgreSQL")
}

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	if _, err := RDB.Ping(context.Background()).Result(); err != nil {
		log.Fatal("❌ Failed to connect to Redis:", err)
	}
	log.Println("✅ Connected to Redis")
}
```

#### 3. File: `internal/utils/crypto.go`
*Menggunakan **Argon2** (Hashing Modern) dan JWT.*

```go
package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

// --- ARGON2 CONFIG ---
type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var p = &argonParams{
	memory:      64 * 1024,
	iterations:  3,
	parallelism: 2,
	saltLength:  16,
	keyLength:   32,
}

// HashPassword creates Argon2id hash
func HashPassword(password string) (string, error) {
	salt := make([]byte, p.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// CheckPassword verifies Argon2id hash
func CheckPassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	var memory, iterations uint32
	var parallelism uint8
	fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	hashToCompare := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(decodedHash)))

	return subtle.ConstantTimeCompare(decodedHash, hashToCompare) == 1
}

// --- JWT & TOKEN UTILS ---

func GenerateAccessToken(userID int64, username string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"name": username,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func GenerateRefreshToken() string {
	return uuid.New().String()
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
```

#### 4. File: `internal/utils/email.go`
*Mengirim Email (Universal: Gmail/Mailtrap).*

```go
package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerificationEmail(toEmail, code string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	
	senderName := os.Getenv("SMTP_SENDER_NAME")
	if senderName == "" {
		senderName = user
	}

	subject := "Subject: Kode Verifikasi Food App\n"
	fromHeader := fmt.Sprintf("From: %s\n", senderName)
	toHeader := fmt.Sprintf("To: %s\n", toEmail)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body := fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; padding: 20px;">
				<div style="background-color: #f4f4f4; padding: 20px; border-radius: 8px;">
					<h2 style="color: #333;">Verifikasi Akun</h2>
					<p>Terima kasih telah mendaftar. Gunakan kode berikut:</p>
					<h1 style="color: #0070f3; background: #fff; padding: 10px; display: inline-block;">%s</h1>
					<p>Kode berlaku selama 15 menit.</p>
				</div>
			</body>
		</html>
	`, code)

	msg := []byte(subject + fromHeader + toHeader + mime + body)
	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", user, password, host)

	return smtp.SendMail(addr, auth, user, []string{toEmail}, msg)
}
```

#### 5. File: `internal/repository/auth_repo.go`
*Query ke Database SQL.*

```go
package repository

import (
	"auth-service/internal/database"
	"auth-service/internal/models"
	"database/sql"
)

func CreateUser(user *models.User) error {
	query := `INSERT INTO users (username, email, password_hash, is_verified) 
              VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return database.DB.QueryRow(query, user.Username, user.Email, user.PasswordHash, user.IsVerified).
		Scan(&user.ID, &user.CreatedAt)
}

func GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, is_verified FROM users WHERE email = $1`
	err := database.DB.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.IsVerified)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UpdateUserVerified(email string) error {
	_, err := database.DB.Exec("UPDATE users SET is_verified = TRUE WHERE email = $1", email)
	return err
}

func CreateRefreshToken(rt models.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (user_id, token_hash, device_id, expires_at, absolute_expires_at, created_at) 
              VALUES ($1, $2, $3, $4, $5, NOW())`
	_, err := database.DB.Exec(query, rt.UserID, rt.TokenHash, rt.DeviceID, rt.ExpiresAt, rt.AbsoluteExpiresAt)
	return err
}

func GetRefreshTokenByHash(hash string) (*models.RefreshToken, error) {
	rt := &models.RefreshToken{}
	query := `SELECT id, user_id, token_hash, expires_at, absolute_expires_at, revoked_at, device_id 
              FROM refresh_tokens WHERE token_hash = $1`
	err := database.DB.QueryRow(query, hash).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.AbsoluteExpiresAt, &rt.RevokedAt, &rt.DeviceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return rt, err
}

func RevokeRefreshToken(id int64) error {
	_, err := database.DB.Exec("UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1", id)
	return err
}

func RevokeAllUserTokens(userID int64) error {
	_, err := database.DB.Exec("UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL", userID)
	return err
}
```

#### 6. File: `internal/service/auth_service.go`
*Logika Bisnis: Register, Login, Verify, Rotate.*

```go
package service

import (
	"auth-service/internal/database"
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/internal/utils"
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func generateOTP() string {
	return strconv.Itoa(100000 + rand.Intn(900000))
}

// 1. REGISTER
func Register(username, email, password string) error {
	if u, _ := repository.GetUserByEmail(email); u != nil {
		return errors.New("email sudah terdaftar")
	}

	hashedPwd, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	newUser := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPwd,
		IsVerified:   false,
	}

	if err := repository.CreateUser(&newUser); err != nil {
		return err
	}

	otp := generateOTP()
	err = database.RDB.Set(context.Background(), "verif:"+email, otp, 15*time.Minute).Err()
	if err != nil {
		return err
	}

	go utils.SendVerificationEmail(email, otp)
	return nil
}

// 2. VERIFY
func VerifyEmail(email, code string) error {
	val, err := database.RDB.Get(context.Background(), "verif:"+email).Result()
	if err == redis.Nil {
		return errors.New("kode verifikasi kadaluarsa")
	}
	if val != code {
		return errors.New("kode salah")
	}

	if err := repository.UpdateUserVerified(email); err != nil {
		return err
	}

	database.RDB.Del(context.Background(), "verif:"+email)
	return nil
}

// 3. LOGIN
func Login(email, password, deviceID string) (string, string, error) {
	user, err := repository.GetUserByEmail(email)
	if err != nil {
		return "", "", errors.New("email atau password salah")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return "", "", errors.New("email atau password salah")
	}

	if !user.IsVerified {
		return "", "", errors.New("akun belum diverifikasi, cek email anda")
	}

	accessToken, _ := utils.GenerateAccessToken(user.ID, user.Username)
	rawRefreshToken := utils.GenerateRefreshToken()

	rt := models.RefreshToken{
		UserID:            user.ID,
		TokenHash:         utils.HashToken(rawRefreshToken),
		DeviceID:          deviceID,
		ExpiresAt:         time.Now().Add(28 * 24 * time.Hour),
		AbsoluteExpiresAt: time.Now().Add(90 * 24 * time.Hour),
	}
	
	if err := repository.CreateRefreshToken(rt); err != nil {
		return "", "", err
	}

	return accessToken, rawRefreshToken, nil
}

// 4. ROTATE REFRESH TOKEN
func RotateRefreshToken(rawToken, deviceID string) (string, string, error) {
	tokenHash := utils.HashToken(rawToken)
	stored, err := repository.GetRefreshTokenByHash(tokenHash)
	
	if err != nil || stored == nil {
		return "", "", errors.New("invalid token")
	}

	// SECURITY: Token Reuse Detection
	if stored.RevokedAt != nil {
		repository.RevokeAllUserTokens(stored.UserID)
		return "", "", errors.New("security alert: token reuse detected")
	}

	if time.Now().After(stored.ExpiresAt) || time.Now().After(stored.AbsoluteExpiresAt) {
		return "", "", errors.New("token expired")
	}

	repository.RevokeRefreshToken(stored.ID)

	newAccess, _ := utils.GenerateAccessToken(stored.UserID, "User") // Ideally fetch username
	newRefresh := utils.GenerateRefreshToken()

	newRt := models.RefreshToken{
		UserID:            stored.UserID,
		TokenHash:         utils.HashToken(newRefresh),
		DeviceID:          stored.DeviceID,
		ExpiresAt:         time.Now().Add(28 * 24 * time.Hour),
		AbsoluteExpiresAt: stored.AbsoluteExpiresAt,
	}
	repository.CreateRefreshToken(newRt)

	return newAccess, newRefresh, nil
}
```

#### 7. File: `internal/handler/auth_handler.go`
*Menangani Request JSON dan Cookie.*

```go
package handler

import (
	"auth-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

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
```

#### 8. File: `internal/middleware/middleware.go`
*Menangani CORS agar Next.js bisa akses.*

```go
package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // URL Frontend Next.js
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Device-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // Wajib TRUE agar cookie dikirim
		MaxAge:           12 * time.Hour,
	})
}
```

#### 9. File: `cmd/api/main.go`
*Entry Point Aplikasi.*

```go
package main

import (
	"auth-service/internal/database"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found")
	}

	// 2. Connect DB
	database.InitDB()
	database.InitRedis()

	// 3. Setup Router
	r := gin.Default()
	r.Use(middleware.CORS())

	// 4. Routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/verify", handler.Verify)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
		auth.POST("/logout", handler.Logout)
	}

	log.Println("🚀 Auth Service running on http://localhost:8080")
	r.Run(":8080")
}
```

#### 10. File: `.env` (Di Root Project)
*Konfigurasi Rahasia. Pilih salah satu settingan SMTP (Gmail / Mailtrap).*

```env
# DATABASE (Sesuaikan dengan PC Murid)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=password_postgres_kamu
DB_NAME=auth_db

# REDIS
REDIS_ADDR=localhost:6379

# JWT KEY (Random string)
JWT_SECRET=rahasia_dapur_bunda_123

# --- PILIH SALAH SATU SMTP DI BAWAH ---

# OPSI 1: MAILTRAP (Untuk Testing Aman)
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USER=username_mailtrap
SMTP_PASS=password_mailtrap
SMTP_SENDER_NAME="Food App Dev <no-reply@dev.com>"

# OPSI 2: GMAIL (Untuk Production)
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USER=email.murid@gmail.com
# SMTP_PASS=app_password_16_digit_dari_google
# SMTP_SENDER_NAME="Food App Official <email.murid@gmail.com>"
```

---

### TAHAP 4: Install & Jalankan

Kembali ke terminal di folder `auth-service`, jalankan perintah "Ajaib" ini untuk mendownload semua library yang dipakai di kode:

```bash
go mod tidy
```
*(Tunggu sampai selesai download)*

Lalu jalankan aplikasi:

```bash
go run cmd/api/main.go
```

Jika muncul tulisan:
```text
✅ Connected to PostgreSQL
✅ Connected to Redis
🚀 Auth Service running on http://localhost:8080
```

---

Test ini akan mensimulasikan user ("robot") yang melakukan: **Daftar -> Cek Redis (ngintip OTP) -> Verifikasi -> Login -> Refresh Token**.

### TAHAP 1: Persiapan Folder Test

1.  Pastikan Anda berada di root folder `auth-service`.
2.  Buat folder baru bernama `tests`.
3.  Di dalam folder `tests`, buat file bernama **`auth_test.go`**.

### TAHAP 2: Copy-Paste Kode Test

File: `tests/auth_test.go`

```go
package tests

import (
	"auth-service/internal/database"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// --- HELPER: Setup Server Virtual ---
func setupRouter() *gin.Engine {
	// 1. Load Env (Mundur satu folder karena file test ada di dalam folder /tests)
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("⚠️  Warning: .env file not found from test directory")
	}

	// 2. Konek Database (Pake database asli local)
	database.InitDB()
	database.InitRedis()

	// 3. Setup Router (Sama persis kayak di main.go)
	gin.SetMode(gin.TestMode) // Supaya log gak berisik
	r := gin.Default()
	r.Use(middleware.CORS())

	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/verify", handler.Verify)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
		auth.POST("/logout", handler.Logout)
	}
	return r
}

// --- HELPER: Bersihkan Data Bekas Test ---
func clearTestData(email string) {
	// Hapus token & user dari Postgres
	// Hapus OTP dari Redis
	database.DB.Exec("DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE email = $1)", email)
	database.DB.Exec("DELETE FROM users WHERE email = $1", email)
	database.RDB.Del(context.Background(), "verif:"+email)
}

// --- TEST UTAMA (END-TO-END) ---
func TestFullAuthFlow(t *testing.T) {
	router := setupRouter()

	// Data Dummy
	email := "robot_test@example.com"
	username := "robot_user"
	password := "passwordRahasia123!"

	// Bersihkan data lama dulu biar gak error "Duplicate"
	clearTestData(email)
	defer clearTestData(email) // Bersihkan lagi setelah selesai

	// --- STEP 1: REGISTER ---
	t.Run("1. Register User Baru", func(t *testing.T) {
		payload := map[string]string{
			"username": username,
			"email":    email,
			"password": password,
		}
		jsonBody, _ := json.Marshal(payload)
		
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code) // Harapannya 201 Created
	})

	// --- STEP 2: AMBIL OTP DARI REDIS (SIMULASI BUKA EMAIL) ---
	var otpCode string
	t.Run("2. Ambil OTP dari Redis", func(t *testing.T) {
		// Kasih jeda dikit biar redis sempat nulis
		time.Sleep(100 * time.Millisecond)

		// Key redis kita tadi: "verif:email"
		val, err := database.RDB.Get(context.Background(), "verif:"+email).Result()
		
		assert.NoError(t, err, "OTP harus ada di Redis")
		assert.NotEmpty(t, val, "OTP tidak boleh kosong")
		
		otpCode = val
		t.Logf("🔑 Kode OTP Ditemukan: %s", otpCode)
	})

	// --- STEP 3: VERIFY EMAIL ---
	t.Run("3. Verifikasi Akun", func(t *testing.T) {
		payload := map[string]string{
			"email": email,
			"code":  otpCode, // Pakai kode yang diambil dari Redis
		}
		jsonBody, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/auth/verify", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code) // Harapannya 200 OK
	})

	// --- STEP 4: LOGIN (DAPAT COOKIE) ---
	var accessToken string
	var refreshCookie *http.Cookie

	t.Run("4. Login & Dapat Token", func(t *testing.T) {
		payload := map[string]string{
			"email":    email,
			"password": password,
		}
		jsonBody, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Cek Body Response (Harus ada access_token)
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		accessToken = resp["access_token"]
		assert.NotEmpty(t, accessToken, "Access Token harus ada")

		// Cek Cookie (Harus ada refresh_token)
		cookies := w.Result().Cookies()
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshCookie = c
			}
		}
		assert.NotNil(t, refreshCookie, "Cookie refresh_token wajib ada")
		assert.True(t, refreshCookie.HttpOnly, "Cookie harus HttpOnly demi keamanan")
	})

	// --- STEP 5: REFRESH TOKEN (ROTATION) ---
	t.Run("5. Refresh Token (Rotation)", func(t *testing.T) {
		// Request kosong, tapi bawa cookie
		req, _ := http.NewRequest("POST", "/auth/refresh", nil)
		req.AddCookie(refreshCookie) // Pasang cookie dari login tadi

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Cek dapat Access Token baru
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		newAT := resp["access_token"]
		
		assert.NotEmpty(t, newAT)
		assert.NotEqual(t, accessToken, newAT, "Token baru harus beda dengan token lama")

		// Cek Cookie di-rotate (Value cookie harus berubah)
		newCookies := w.Result().Cookies()
		var newRefreshCookie *http.Cookie
		for _, c := range newCookies {
			if c.Name == "refresh_token" {
				newRefreshCookie = c
			}
		}
		assert.NotEqual(t, refreshCookie.Value, newRefreshCookie.Value, "Refresh token hash harus berubah (Rotated)")
	})
}
```

### TAHAP 3: Jalankan Test

Buka terminal di root project `auth-service`, lalu jalankan:

1.  **Download Library Testify** (Wajib, untuk fungsi `assert`)
    ```bash
    go get github.com/stretchr/testify
    go mod tidy
    ```

2.  **Jalankan Test**
    ```bash
    go test ./tests/... -v
    ```

### Hasil yang Diharapkan

Jika semua lancar, terminal akan menampilkan output hijau seperti ini:

```text
=== RUN   TestFullAuthFlow
=== RUN   TestFullAuthFlow/1._Register_User_Baru
=== RUN   TestFullAuthFlow/2._Ambil_OTP_dari_Redis
    auth_test.go:88: 🔑 Kode OTP Ditemukan: 538192
=== RUN   TestFullAuthFlow/3._Verifikasi_Akun
=== RUN   TestFullAuthFlow/4._Login_&_Dapat_Token
=== RUN   TestFullAuthFlow/5._Refresh_Token_(Rotation)
--- PASS: TestFullAuthFlow (0.24s)
PASS
ok      auth-service/tests      0.552s
```

Ini membuktikan backend Anda sudah **100% Berfungsi** dari Database, Redis, Logic, hingga Security-nya sebelum disentuh oleh Frontend.