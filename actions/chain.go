package actions

import "tastien.com/chat-bot/bot"

var chain bot.ActionChain

func GetActionChain() bot.ActionChain {
	if chain == nil {
		chain = bot.ActionChain{
			&ProcessedMessageAction{}, // 避免重复处理消息
			&TextMessageAction{},      // 处理文字消息
			&UnknownMessageAction{},   // 兜底的消息处理
		}
	}
	return chain
}
