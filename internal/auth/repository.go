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
	rdb *redis.Client // used only for OTP storage
}

type IAuthRepository interface {
	StoreOTP(email string, otp string, expiration time.Duration) error
	VerifyOTP(email string, otp string) (bool, error)
	// CreateSession persists the session and returns the IDs of any sessions
	// that were evicted to enforce the max-2-per-user limit.
	CreateSession(session *models.Session) (evictedIDs []string, err error)
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
		// Delete OTP key immediately after successful verification
		r.rdb.Del(ctx, key)
		return true, nil
	}

	return false, nil
}

// CreateSession persists a new session to PostgreSQL. If the user already has
// 2 active sessions, the oldest one(s) are deleted first. The IDs of any
// evicted sessions are returned so the caller can invalidate their cache entries.
func (r *AuthRepository) CreateSession(session *models.Session) (evictedIDs []string, err error) {
	// Enforce the concurrency limit: max 2 active sessions per user
	var activeSessions []models.Session
	if err = r.db.Where("user_id = ? AND expires_at > ?", session.UserID, time.Now()).
		Order("created_at ASC").
		Find(&activeSessions).Error; err != nil {
		return nil, err
	}

	if len(activeSessions) >= 2 {
		toEvict := len(activeSessions) - 2 + 1
		evictedIDs = make([]string, 0, toEvict)
		for i := range toEvict {
			if dbErr := r.db.Where("id = ?", activeSessions[i].ID).Delete(&models.Session{}).Error; dbErr == nil {
				evictedIDs = append(evictedIDs, activeSessions[i].ID)
			}
		}
	}

	return evictedIDs, r.db.Create(session).Error
}

// DeleteSession removes the session from PostgreSQL only.
func (r *AuthRepository) DeleteSession(sessionID string) error {
	return r.db.Where("id = ?", sessionID).Delete(&models.Session{}).Error
}
