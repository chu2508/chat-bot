package bot

import (
	"context"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	recontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/sashabaranov/go-openai"
	"tastien.com/chat-bot/cache"
)

type Bot struct {
	Lark         *lark.Client
	GPT          *openai.Client
	SessionCache cache.SessionCacheInterface
	MessageCache cache.MessageCacheInterface
	Actions      ActionChain
	Config       *Config
}

func NewBot(cfg *Config, actions ActionChain) *Bot {
	gptConfig := openai.DefaultConfig(cfg.OpenAIKeys[0])
	gptConfig.BaseURL = cfg.OpenAIUrl + "/v1"

	return &Bot{
		Lark:         lark.NewClient(cfg.FeishuAppId, cfg.FeishuAppSecret),
		GPT:          openai.NewClientWithConfig(gptConfig),
		SessionCache: cache.GetSessionCache(),
		MessageCache: cache.GetMessageCache(),
		Actions:      actions,
		Config:       cfg,
	}
}

func (b *Bot) HandleReceive(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	info := NewMsgInfo(event)

	b.Actions.Execute(ctx, b, info)
	return nil
}

func (b *Bot) HandleCardAction(ctx context.Context, event *larkcard.CardAction) (interface{}, error) {

	return nil, nil
}

// 实现用户加群欢迎语
func (b *Bot) HandleUserAdded(ctx context.Context, event *larkim.P2ChatMemberUserAddedV1) error {
	// 先获取用户信息
	userId := event.Event.Users[0].UserId.UserId
	fmt.Println("HandleUserAdded: ", userId)

	req := recontact.NewGetUserReqBuilder().UserId(*userId).UserIdType("user_id").Build()
	res, err := b.Lark.Contact.User.Get(ctx, req)
	if err != nil {
		fmt.Println("GetUserError:", err)
		return err
	}

	// 获取欢迎语
	greetStr, err := b.getGreetText(ctx, res.Data.User, event)
	if err != nil {
		return err
	}
	// 在发送群消息
	return b.sendMsg(ctx, *event.Event.ChatId, greetStr)
}

func (b *Bot) getGreetText(ctx context.Context, user *recontact.User, event *larkim.P2ChatMemberUserAddedV1) (string, error) {
	// 根据用户信息里的名称和职位生成欢迎语
	userName := user.Name
	userJobTitle := user.JobTitle
	fmt.Println("UserName: ", userName)
	fmt.Println("JobTitle: ", userJobTitle)
	fmt.Println("UserCustomData: ", user.CustomAttrs)
	req, err := b.GPT.CreateCompletion(ctx, openai.CompletionRequest{
		Model:       openai.GPT3TextDavinci003,
		Prompt:      fmt.Sprintf("写一个100字的欢迎语，欢迎%s加入%s，他的职位是%s，欢迎语需要活泼有趣。", *userName, *event.Event.Name, *userJobTitle),
		Temperature: 1,
		Stream:      false,
	})

	if err != nil {
		return "", err
	}

	return req.Choices[0].Text, nil
}

func (b *Bot) sendMsg(ctx context.Context, chatId string, messageText string) error {
	body, err := larkim.NewCreateMessagePathReqBodyBuilder().MsgType(larkim.MsgTypeText).
		Content(larkim.NewTextMsgBuilder().Text(messageText).Build()).Build()
	if err != nil {
		fmt.Printf("SendMsgError: %s", err)
		return err
	}

	req := larkim.NewCreateMessageReqBuilder().Body(body).ReceiveIdType("chat_id").Build()
	_, err = b.Lark.Im.Message.Create(ctx, req)
	if err != nil {
		fmt.Printf("SendMsgError: %s", err)
		return err
	}
	return nil
}
