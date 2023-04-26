package actions

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/sashabaranov/go-openai"
	"tastien.com/chat-bot/bot"
	"tastien.com/chat-bot/cache"
	"tastien.com/chat-bot/utils"
)

// 处理文字消息
type TextMessageAction struct {
}

func (t *TextMessageAction) Execute(payload *bot.ActionPayload) (bool, error) {
	if payload.Info.MsgType != "text" {
		return true, nil
	}
	mode := payload.Bot.SessionCache.GetMode(payload.Info.SessionId)
	if mode == cache.SessionModeCreateImage {
		gpt := payload.Bot.GPT
		req := openai.ImageRequest{
			Prompt:         payload.Info.Content,
			Size:           openai.CreateImageSize1024x1024,
			ResponseFormat: openai.CreateImageResponseFormatB64JSON,
			N:              1,
		}
		res, err := gpt.CreateImage(payload.Ctx, req)
		if err != nil {
			fmt.Println("create image error: ", err)
			return false, err
		}

		img, err := uploadImage(payload.Bot, res.Data[0].B64JSON)
		if err != nil {
			fmt.Println("upload image error: ", err)
			return false, err
		}
		err = replyImage(payload, img)

		return false, err
	}

	message, err := doPrecess(payload)
	if err != nil {
		fmt.Println("get chat message error: ", err)
		return false, err
	}
	message, err = utils.ProcessMessage(message)
	if err != nil {
		fmt.Println("processMessage error: ", err)
		return false, err
	}
	_, err = replyTextMessage(payload, message)
	return false, err
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
	if _, isCreateImage := utils.EitherCutPrefix(content, "/image", "生成图片"); isCreateImage {
		payload.Bot.SessionCache.Clear(sessionId)
		payload.Bot.SessionCache.SetMode(sessionId, cache.SessionModeCreateImage)
		return "🤖️：已开启图片生成模式，请回复这条消息，生成图片。", nil
	}
	if msg, isCosplay := utils.EitherCutPrefix(content, "/cosplay", "角色扮演"); isCosplay {
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
	if _, isClear := utils.EitherCutPrefix(content, "/clear", "清除"); isClear {
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
		Model:       payload.Bot.Config.OpenAIModel,
		Temperature: 0.6,
	}
	res, err := gpt.CreateChatCompletion(payload.Ctx, req)
	if err != nil {
		fmt.Println("gpt3 error:", err)
		return err.Error(), err
	}
	messages = append(messages, res.Choices[0].Message)
	payload.Bot.SessionCache.SetMode(sessionId, cache.SessionModeChat)
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

func uploadImage(bot *bot.Bot, base64Str string) (*string, error) {
	imageBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	client := bot.Lark
	resp, err := client.Im.Image.Create(context.Background(),
		larkim.NewCreateImageReqBuilder().
			Body(larkim.NewCreateImageReqBodyBuilder().
				ImageType(larkim.ImageTypeMessage).
				Image(bytes.NewReader(imageBytes)).
				Build()).
			Build())

	// 处理错误
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, err
	}
	return resp.Data.ImageKey, nil
}
func replyImage(payload *bot.ActionPayload, ImageKey *string) error {
	//fmt.Println("sendMsg", ImageKey, msgId)
	bot := payload.Bot
	ctx := payload.Ctx

	msgImage := larkim.MessageImage{ImageKey: *ImageKey}
	content, err := msgImage.String()
	if err != nil {
		fmt.Println(err)
		return err
	}
	client := bot.Lark

	resp, err := client.Im.Message.Reply(ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(payload.Info.MessageId).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeImage).
			Uuid(uuid.New().String()).
			Content(content).
			Build()).
		Build())

	// 处理错误
	if err != nil {
		fmt.Println(err)
		return err
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return err
	}
	return nil
}
