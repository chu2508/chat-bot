package conditions

import "tastien.com/chat-bot/bot"

// 避免重复处理消息
type ProcessedMessageAction struct {
}

func (p *ProcessedMessageAction) Execute(action *bot.ActionPayload) (bool, error) {
	if action.Bot.MessageCache.HasMessage(action.Info.MessageId) {
		return false, nil
	}
	action.Bot.MessageCache.SetMessage(action.Info.MessageId)
	return true, nil
}

// 判断是否支持处理这个消息
type SupportedMessageAction struct {
}

func (*SupportedMessageAction) Execute(payload *bot.ActionPayload) (bool, error) {
	if payload.Info.HandlerType == bot.PersonalHandler {
		return true, nil
	}
	if payload.Info.HandlerType == bot.GroupHandler {
		return true, nil
	}

	return false, nil
}
