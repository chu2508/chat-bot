package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type messageCache struct {
	cache *cache.Cache
}

type MessageCacheInterface interface {
	HasMessage(msgId string) bool
	SetMessage(msgId string)
}

func (m *messageCache) HasMessage(msgId string) bool {
	if _, ok := m.cache.Get(msgId); ok {
		return true
	}

	return false
}

func (m *messageCache) SetMessage(msgId string) {
	m.cache.Set(msgId, true, maxCacheTime)
}

func GetMessageCache() MessageCacheInterface {
	return &messageCache{
		cache: cache.New(maxCacheTime, time.Hour),
	}
}
