package actions

import (
	"tastien.com/chat-bot/actions/conditions"
	"tastien.com/chat-bot/bot"
)

var chain bot.ActionChain

func GetActionChain() bot.ActionChain {
	if chain == nil {
		chain = bot.ActionChain{
			&conditions.ProcessedMessageAction{}, // 避免重复处理消息
			&conditions.AtMessageAction{},        // 判断是否@机器人
			&conditions.SupportedMessageAction{}, // 处理支持的消息
			&TextMessageAction{},                 // 处理文字消息
			&UnknownMessageAction{},              // 兜底的消息处理
		}
	}
	return chain
}
