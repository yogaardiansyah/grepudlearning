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