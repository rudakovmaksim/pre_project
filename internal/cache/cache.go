package cases

// this need for modification and testing

import (
	"strings"
	"time"

	"project/internal/entities"
)

type CacheCoin struct {
	Cache    map[string]data
	cacheKey string
}

type data struct {
	coin    []entities.Coin
	expires time.Time
}

func NewCacheCoin() CacheCoin {
	return CacheCoin{
		Cache:    make(map[string]data, 50),
		cacheKey: "",
	}
}

func (c *CacheCoin) FindKeyGetCoin(metodName string, titles []string) ([]entities.Coin, bool) {

	c.cacheKey = c.generateCacheKey(metodName, titles)

	if cached, ok := c.Cache[c.cacheKey]; ok && time.Now().Before(cached.expires) {
		return cached.coin, true
	}

	return nil, false
}

func (c *CacheCoin) SetCoin(updateCoins []entities.Coin, timeLimit int) {
	c.Cache[c.cacheKey] = data{
		coin:    updateCoins,
		expires: time.Now().Add(time.Duration(timeLimit) * time.Minute),
	}
}

func (c *CacheCoin) generateCacheKey(metod string, titles []string) string {
	return metod + ":" + strings.Join(titles, ",")
}
