package auth

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vivek6201/biolynq/internal/models"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

type IAuthRepository interface {
	StoreOTP(email string, otp string, expiration time.Duration) error
	VerifyOTP(email string, otp string) (bool, error)
	CreateSession(session *models.Session) error
	DeleteSession(sessionID string) error
}

func NewAuthRepository(db *gorm.DB, rdb *redis.Client) IAuthRepository {
	return &AuthRepository{
		db:  db,
		rdb: rdb,
	}
}

func (r *AuthRepository) StoreOTP(email string, otp string, expiration time.Duration) error {
	ctx := context.Background()
	key := "otp:" + email
	return r.rdb.Set(ctx, key, otp, expiration).Err()
}

func (r *AuthRepository) VerifyOTP(email string, otp string) (bool, error) {
	ctx := context.Background()
	key := "otp:" + email
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	if val == otp {
		// Delete key immediately after successful verification
		r.rdb.Del(ctx, key)
		return true, nil
	}

	return false, nil
}

func (r *AuthRepository) CreateSession(session *models.Session) error {
	// Enforce the concurrency limit: max 2 active sessions per user
	var activeSessions []models.Session
	err := r.db.Where("user_id = ? AND expires_at > ?", session.UserID, time.Now()).
		Order("created_at ASC").
		Find(&activeSessions).Error

	if err == nil && len(activeSessions) >= 2 {
		// Evict the oldest active sessions to bring the count under 2
		sessionsToDelete := len(activeSessions) - 2 + 1
		for i := range sessionsToDelete {
			r.db.Where("id = ?", activeSessions[i].ID).Delete(&models.Session{})
		}
	}

	return r.db.Create(session).Error
}

func (r *AuthRepository) DeleteSession(sessionID string) error {
	return r.db.Where("id = ?", sessionID).Delete(&models.Session{}).Error
}
