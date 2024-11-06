package telegrambot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type MessageType string

// 文本消息
type Message struct {
	MessageTimestamp int64       //消息时间戳秒
	FromUserId       int64       //发出消息的用户id
	FromGroupId      int64       //群组id
	Msg              string      //text消息
	MsgType          MessageType //消息类型
	IsCallBack       bool        //是否是回调消息
	MessageId        int
	Base             TgMessage
}

const (
	TypeAll        MessageType = "all"     //所有
	TypePrivate    MessageType = "private" //私聊
	TypeGroup      MessageType = "group"
	TypeSuperGroup MessageType = "supergroup" //超级群聊
	TypeChannel    MessageType = "channel"
)

// 处理消息的
type MessageHandler interface {
	Type() []MessageType
	//收到的消息(简化只处理文本消息)
	ReciveMsg(Bot, MessageType, *Message)
	//所有未经过滤的消息
	AllMsg(Bot, TgMessage)
}

type Bot interface {
	Run() error
	Stop()
	SendMsg(tgbotapi.Chattable) error
	RegHandler(MessageHandler)
}

func NewBot(token string, isdebug bool) (Bot, error) {
	b := &gbot{}
	err := b.init(token, isdebug)
	if err != nil {
		return nil, err
	}
	return b, nil
}
