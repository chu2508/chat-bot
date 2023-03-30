package actions

import (
	"fmt"
	"time"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/sashabaranov/go-openai"
	"tastien.com/chat-bot/bot"
)

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

// 处理文字消息
type TextMessageAction struct {
}

func (t *TextMessageAction) Execute(payload *bot.ActionPayload) (bool, error) {
	if payload.Info.MsgType != "text" {
		return true, nil
	}
	message, err := doPrecess(payload)
	if err != nil {
		return false, err
	}
	_, err = replyTextMessage(payload, message)
	return false, err
}

func doPrecess(payload *bot.ActionPayload) (string, error) {
	gpt := payload.Bot.GPT
	sessionId := payload.Info.SessionId
	messages := payload.Bot.SessionCache.GetMessage(sessionId)
	if messages == nil {
		messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are ChatGPT, a large language model trained by OpenAI. Answer as concisely as possible.\nKnowledge cutoff: 2021-09-01\nCurrent date: " + time.Now().Format("2006-01-02"),
			},
		}
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: payload.Info.Content,
	})

	req := openai.ChatCompletionRequest{
		Messages: messages,
		Model:    openai.GPT3Dot5Turbo,
	}
	res, err := gpt.CreateChatCompletion(payload.Ctx, req)
	if err != nil {
		fmt.Println("gpt3 error:", err)
		return "", err
	}
	return res.Choices[0].Message.Content, nil
}

func replyTextMessage(payload *bot.ActionPayload, replayMessage string) (*larkim.ReplyMessageResp, error) {
	lark := payload.Bot.Lark
	content := larkim.NewTextMsgBuilder().
		Text(replayMessage).
		Build()
	body := larkim.NewReplyMessageReqBodyBuilder().
		Content(content).
		MsgType("text").
		Build()
	req := larkim.NewReplyMessageReqBuilder().
		Body(body).
		MessageId(payload.Info.MessageId).
		Build()
	res, err := lark.Im.Message.Reply(payload.Ctx, req)
	fmt.Println("reply res: ", res)

	return res, err
}

// 未知消息类型处理
type UnknownMessageAction struct {
}

func (u *UnknownMessageAction) Execute(payload *bot.ActionPayload) (bool, error) {
	_, err := replyTextMessage(payload, "🤖️：还不支持的消息类型，敬请期待功能开发！")
	return false, err
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
