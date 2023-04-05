package bot

import "context"

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
