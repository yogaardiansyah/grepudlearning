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