package bot

import (
	"encoding/json"
	"fmt"
	"regexp"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type MsgInfo struct {
	HandlerType HandlerType
	MsgType     string
	MessageId   string
	ChatId      string
	Content     string
	SessionId   string
}

type HandlerType string

const (
	GroupHandler    HandlerType = "group"
	PersonalHandler HandlerType = "personal"
)

func NewMsgInfo(event *larkim.P2MessageReceiveV1) *MsgInfo {
	msg := event.Event.Message
	messageId := msg.MessageId
	rootId := msg.RootId
	chatId := msg.ChatId
	content := msg.Content
	msgType := msg.MessageType

	// 获取sessionId，用于后续的回复，如果有rootId，则使用rootId，否则使用messageId
	sessionId := rootId
	if sessionId == nil || *sessionId == "" {
		sessionId = messageId
	}

	return &MsgInfo{
		ChatId:    *chatId,
		MessageId: *messageId,
		SessionId: *sessionId,
		MsgType:   *msgType,
		Content:   parseContent(*content),
	}
}

func msgFilter(msg string) string {
	//replace @到下一个非空的字段 为 ''
	regex := regexp.MustCompile(`@[^ ]*`)
	return regex.ReplaceAllString(msg, "")

}
func parseContent(content string) string {
	//"{\"text\":\"@_user_1  hahaha\"}",
	//only get text content hahaha
	var contentMap map[string]interface{}
	err := json.Unmarshal([]byte(content), &contentMap)
	if err != nil {
		fmt.Println(err)
	}
	if contentMap["text"] == nil {
		return ""
	}
	text := contentMap["text"].(string)
	return msgFilter(text)
}
