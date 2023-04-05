package bot

import (
	"context"

	lark "github.com/larksuite/oapi-sdk-go/v3"
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
