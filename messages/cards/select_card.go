package cards

import (
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	"github.com/samber/lo"
	"tastien.com/chat-bot/bot"
)

type Menu struct {
	Label string
	Value string
}

func newCard(header *larkcard.MessageCardHeader, elements ...larkcard.MessageCardElement) *larkcard.MessageCard {
	config := larkcard.NewMessageCardConfig().
		WideScreenMode(false).
		EnableForward(true).
		UpdateMulti(false).
		Build()
	return larkcard.NewMessageCard().
		Config(config).
		Header(header).
		Elements(elements).
		Build()
}

var roles = []string{
	"role1",
	"role2",
}

func NewRolesCard(info *bot.MsgInfo) (string, error) {
	header := newCardHeader("🎭 角色列表")
	desc := larkcard.NewMessageCardDiv().Text(newPlainText("提醒：选择内置角色，快速进入角色扮演模式。")).Build()
	values := map[string]interface{}{
		"value":     "",
		"sessionId": info.SessionId,
		"messageId": info.SessionId,
	}
	menus := lo.Map(roles, func(role string, idx int) *Menu {
		return &Menu{
			Label: role,
			Value: role,
		}
	})

	return newCard(header, desc, newMenu("请选择角色", values, menus)).String()
}

func newMenu(placeholder string, value map[string]interface{}, menus []*Menu) *larkcard.MessageCardEmbedSelectMenuStatic {
	var options []*larkcard.MessageCardEmbedSelectOption
	for _, menu := range menus {
		options = append(options, larkcard.NewMessageCardEmbedSelectOption().Text(newPlainText(menu.Label)).Value(menu.Value).Build())
	}

	return larkcard.NewMessageCardEmbedSelectMenuStatic().
		MessageCardEmbedSelectMenuStatic(larkcard.NewMessageCardEmbedSelectMenuBase().
			Options(options).
			Placeholder(newPlainText(placeholder)).
			Value(value).
			Build()).
		Build()
}

func newCardHeader(title string) *larkcard.MessageCardHeader {
	return larkcard.NewMessageCardHeader().Title(newPlainText(title)).Build()
}

func newPlainText(text string) *larkcard.MessageCardPlainText {
	return larkcard.NewMessageCardPlainText().Content(text).Build()
}
