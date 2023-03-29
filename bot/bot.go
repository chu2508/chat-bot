package bot

import (
	"context"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/sashabaranov/go-openai"
	"tastien.com/chat-bot/cache"
)

type ActionPayload struct {
	Ctx  context.Context
	Bot  *Bot
	Info *MsgInfo
}

// 处理动作接口，返回true表示继续处理，false表示不再处理，error表示处理出错
type Action interface {
	Execute(payload *ActionPayload) (bool, error)
}

// 动作链，按顺序执行，如果有一个动作返回false或者error，则不再执行后续动作
type ActionChain []Action

func (a ActionChain) Execute(ctx context.Context, bot *Bot, info *MsgInfo) (bool, error) {
	for _, action := range a {
		if ok, err := action.Execute(&ActionPayload{
			Ctx:  ctx,
			Bot:  bot,
			Info: info,
		}); !ok || err != nil {
			return false, err
		}
	}
	return true, nil
}

type Bot struct {
	Lark         *lark.Client
	GPT          *openai.Client
	SessionCache cache.SessionCacheInterface
	MessageCache cache.MessageCacheInterface
	Actions      ActionChain
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
	}
}

func (b *Bot) HandleReceive(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	info := NewMsgInfo(event)

	b.Actions.Execute(ctx, b, info)
	return nil
}
