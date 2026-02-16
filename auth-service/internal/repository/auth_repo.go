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

func GetUserByID(id int64) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email FROM users WHERE id = $1`
	err := database.DB.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}