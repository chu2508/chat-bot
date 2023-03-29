package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sashabaranov/go-openai"
)

const (
	maxLength    = 4096
	maxCacheTime = time.Hour * 12
)

type SessionMeta struct {
	Messages []openai.ChatCompletionMessage `json:"messages,omitempty"`
}

type sessionCache struct {
	cache *cache.Cache
}

type SessionCacheInterface interface {
	GetMessage(sessionId string) []openai.ChatCompletionMessage
	SetMessage(sessionId string, message []openai.ChatCompletionMessage)
	Clear(sessionId string)
}

var _sessionCache SessionCacheInterface

func GetSessionCache() SessionCacheInterface {
	if _sessionCache == nil {
		_sessionCache = &sessionCache{
			cache: cache.New(maxCacheTime, time.Hour),
		}
	}

	return _sessionCache
}

// GetMessage 获取消息, 如果没有则返回nil
func (s *sessionCache) GetMessage(sessionId string) []openai.ChatCompletionMessage {
	if v, ok := s.cache.Get(sessionId); ok {
		return v.(SessionMeta).Messages
	}

	return nil
}

// SetMessage 设置消息, 如果超过长度则删除最早的消息, 默认缓存时间为12小时
func (s *sessionCache) SetMessage(sessionId string, message []openai.ChatCompletionMessage) {
	// 限制消息长度如果超过限制则删除最早的消息
	for getMessagesTotalLength(message) > maxLength {
		message = append(message[:1], message[2:]...)
	}

	// 从cache 中获取Session，如果没有则创建一个新的
	var session SessionMeta
	if v, ok := s.cache.Get(sessionId); !ok {
		session = SessionMeta{
			Messages: message,
		}
	} else {
		session = v.(SessionMeta)
		session.Messages = message
	}
	s.cache.Set(sessionId, session, maxCacheTime)
}

// Clear 清除缓存
func (s *sessionCache) Clear(sessionId string) {
	s.cache.Delete(sessionId)
}

func getMessagesTotalLength(messages []openai.ChatCompletionMessage) int {
	length := 0
	for _, msg := range messages {
		length += len(msg.Content)
	}

	return length
}
