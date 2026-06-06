package links

import (
	"github.com/redis/go-redis/v9"
	"github.com/vivek6201/biolynq/internal/cache"
)

// Caches holds all typed Redis caches used by the links domain.
type Caches struct {
	Links *cache.Cache[[]LinkResponse] // list of all links for a profile
	Link  *cache.Cache[LinkResponse]   // single link by ID
}

// NewCaches constructs all link-domain caches from a single Redis client.
func NewCaches(rdb *redis.Client) *Caches {
	return &Caches{
		Links: cache.NewCache[[]LinkResponse](rdb),
		Link:  cache.NewCache[LinkResponse](rdb),
	}
}
