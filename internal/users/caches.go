package users

import (
	"github.com/redis/go-redis/v9"
	"github.com/vivek6201/biolynq/internal/cache"
	"github.com/vivek6201/biolynq/internal/models"
)

// Caches holds all typed Redis caches used by the users domain.
type Caches struct {
	Session *cache.Cache[models.Session]
	Profile *cache.Cache[models.Profile]
	User    *cache.Cache[models.User]
}

// NewCaches constructs all user-domain caches from a single Redis client.
func NewCaches(rdb *redis.Client) *Caches {
	return &Caches{
		Session: cache.NewCache[models.Session](rdb),
		Profile: cache.NewCache[models.Profile](rdb),
		User:    cache.NewCache[models.User](rdb),
	}
}
