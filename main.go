package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	sdkginext "github.com/larksuite/oapi-sdk-gin"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	"github.com/spf13/pflag"
	"tastien.com/chat-bot/actions"
	"tastien.com/chat-bot/bot"
)

var (
	cfgPath = pflag.StringP("config", "c", "./config.yaml", "apiserver config file path.")
)

func main() {
	cfg := bot.LoadConfig(*cfgPath)

	run, _ := CreateApp(cfg)
	run()
}

func CreateApp(cfg *bot.Config) (func(), *gin.Engine) {
	bot := bot.NewBot(cfg, actions.GetActionChain())

	eventHandler := dispatcher.NewEventDispatcher(cfg.FeishuAppVerificationToken, cfg.FeishuAppEncryptKey).
		OnP2MessageReceiveV1(bot.HandleReceive).
		OnP2ChatMemberUserAddedV1(bot.HandleUserAdded)

	cardHandler := larkcard.NewCardActionHandler(cfg.FeishuAppVerificationToken,
		cfg.FeishuAppEncryptKey,
		bot.HandleCardAction,
	)

	r := gin.Default()
	// 定义一个 GET 路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	r.POST("/webhook/event", sdkginext.NewEventHandlerFunc(eventHandler))
	r.POST("/webhook/card", sdkginext.NewCardActionHandlerFunc(cardHandler))

	return func() {
		r.Run(fmt.Sprintf(":%d", cfg.HttpPort))
	}, r
}

// func replyMsg(msg string, repliedMsgId string) error {
// 	larkClient := lark.NewClient("cli_a4a878685179d013", "3kcDGhCCTe3ha1nI9Wo0jbS3muhMvV1m") // 默认配置为自建应用
// 	content := larkim.NewTextMsgBuilder().
// 		Text(msg).
// 		Build()

// 	resp, err := larkClient.Im.Message.Reply(context.Background(), larkim.NewReplyMessageReqBuilder().
// 		MessageId(repliedMsgId).Body(larkim.NewReplyMessageReqBodyBuilder().
// 		MsgType(larkim.MsgTypeText).
// 		Uuid(uuid.New().String()).
// 		Content(content).
// 		Build()).
// 		Build())

// 	// 处理错误
// 	if err != nil {
// 		fmt.Println(err)
// 		return err
// 	}

// 	// 服务端错误处理
// 	if !resp.Success() {
// 		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
// 		return err
// 	}
// 	return nil
// }
