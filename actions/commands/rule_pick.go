package commands

import (
	"tastien.com/chat-bot/bot"
	"tastien.com/chat-bot/utils"
)

type RulePickAction struct{}

func (r *RulePickAction) Execute(payload *bot.ActionPayload) (bool, error) {
	if _, isRules := utils.EitherCutPrefix(payload.Info.MessageId, "/rules", "角色列表"); isRules {
		return false, nil
	}

	return true, nil
}
