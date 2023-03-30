package actions

import (
	"encoding/json"
	"fmt"
	"strings"
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
		fmt.Println("get chat message error: ", err)
		return false, err
	}
	message, err = processMessage(message)
	if err != nil {
		fmt.Println("processMessage error: ", err)
		return false, err
	}
	_, err = replyTextMessage(payload, message)
	return false, err
}

func processMessage(msg interface{}) (string, error) {
	msg = strings.TrimSpace(msg.(string))
	msgB, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	msgStr := string(msgB)

	if len(msgStr) >= 2 {
		msgStr = msgStr[1 : len(msgStr)-1]
	}
	return msgStr, nil
}

var defaultPrompt = openai.ChatCompletionMessage{
	Role:    openai.ChatMessageRoleSystem,
	Content: "You are ChatGPT, a large language model trained by OpenAI. Answer as concisely as possible.\nKnowledge cutoff: 2021-09-01\nCurrent date: " + time.Now().Format("2006-01-02"),
}

func doPrecess(payload *bot.ActionPayload) (string, error) {
	gpt := payload.Bot.GPT
	sessionId := payload.Info.SessionId
	messages := payload.Bot.SessionCache.GetMessage(sessionId)
	content := payload.Info.Content
	fmt.Println("user message content: ", payload.Info.Content)
	fmt.Println("session messages: ", messages)
	if msg, isCosplay := eitherCutPrefix(content, "/cosplay", "角色扮演"); isCosplay {
		payload.Bot.SessionCache.Clear(sessionId)
		messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: msg,
			},
		}
		payload.Bot.SessionCache.SetMessage(sessionId, messages)
		return "🤖️：已开启角色扮演模式，请回复这条消息，开始你的表演。", nil
	}
	if _, isClear := eitherCutPrefix(content, "/clear", "清除"); isClear {
		messages := payload.Bot.SessionCache.GetMessage(sessionId)
		if messages == nil {
			messages = []openai.ChatCompletionMessage{defaultPrompt}
		} else {
			messages = messages[:1]
		}
		payload.Bot.SessionCache.Clear(sessionId)
		payload.Bot.SessionCache.SetMessage(sessionId, messages)
		return "🤖️：已清除会话上下文信息。", nil
	}
	if messages == nil {

		messages = []openai.ChatCompletionMessage{defaultPrompt}
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: payload.Info.Content,
	})

	req := openai.ChatCompletionRequest{
		Messages:    messages,
		Model:       openai.GPT3Dot5Turbo,
		Temperature: 0.6,
	}
	res, err := gpt.CreateChatCompletion(payload.Ctx, req)
	if err != nil {
		fmt.Println("gpt3 error:", err)
		return "", err
	}
	messages = append(messages, res.Choices[0].Message)
	payload.Bot.SessionCache.SetMessage(sessionId, messages)
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
	fmt.Println("reply msg: ", replayMessage)
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

func eitherCutPrefix(s string, prefix ...string) (string, bool) {
	// 任一前缀匹配则返回剩余部分
	for _, p := range prefix {
		if strings.HasPrefix(s, p) {
			return strings.TrimPrefix(s, p), true
		}
	}
	return s, false
}
