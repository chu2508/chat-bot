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

type SessionMode string

const (
	SessionModeChat        SessionMode = "chat"
	SessionModeCreateImage SessionMode = "create_image"
)

type SessionMeta struct {
	Messages []openai.ChatCompletionMessage `json:"messages,omitempty"`
	Mode     SessionMode                    `json:"mode,omitempty"`
}

type sessionCache struct {
	cache *cache.Cache
}

type SessionCacheInterface interface {
	GetMessage(sessionId string) []openai.ChatCompletionMessage
	SetMessage(sessionId string, message []openai.ChatCompletionMessage)
	SetMode(sessionId string, mode SessionMode)
	GetMode(sessionId string) SessionMode
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
	var session SessionMeta = s.getSession(sessionId)
	session.Messages = message
	s.cache.Set(sessionId, session, maxCacheTime)
}

// Clear 清除缓存
func (s *sessionCache) Clear(sessionId string) {
	s.cache.Delete(sessionId)
}

func (s *sessionCache) SetMode(sessionId string, mode SessionMode) {
	var session SessionMeta = s.getSession(sessionId)
	session.Mode = mode
	s.cache.Set(sessionId, session, maxCacheTime)
}
func (s *sessionCache) GetMode(sessionId string) SessionMode {
	if v, ok := s.cache.Get(sessionId); ok {
		return v.(SessionMeta).Mode
	}

	return ""
}

func (s *sessionCache) getSession(sessionId string) SessionMeta {
	if v, ok := s.cache.Get(sessionId); ok {
		return v.(SessionMeta)
	}

	return SessionMeta{}
}

func getMessagesTotalLength(messages []openai.ChatCompletionMessage) int {
	length := 0
	for _, msg := range messages {
		length += len(msg.Content)
	}

	return length
}
