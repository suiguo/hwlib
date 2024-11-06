package telegrambot

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type TgMessage tgbotapi.Update
type Status int32

const (
	StatusStop Status = iota
	StatusRun
)

type gbot struct {
	isdebug bool
	*zap.SugaredLogger
	token   string
	api     *tgbotapi.BotAPI
	data_ch tgbotapi.UpdatesChannel //消息channel
	close   context.CancelFunc
	ctx     context.Context
	status  atomic.Int32
	MessageHandler
}

func (g *gbot) initlog(debug bool) {
	writeSyncer := zapcore.AddSync(os.Stdout)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	cores := make([]zapcore.Core, 0)
	if debug {
		cores = append(cores, zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel))
	} else {
		cores = append(cores, zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel))
	}
	handler := zapcore.NewTee(cores...)
	zaplogger := zap.New(handler, zap.AddCaller(), zap.AddCallerSkip(0)) //修改堆栈深度
	sugarLogger := zaplogger.Sugar()
	g.SugaredLogger = sugarLogger
}
func (g *gbot) init(token string, debug bool) (e error) {
	g.initlog(debug)
	defer func() {
		if e != nil && g.SugaredLogger != nil {
			g.Debugw("init", "err", e)
		}
	}()
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}
	if api == nil {
		return fmt.Errorf("api is nil")
	}
	g.api = api
	g.token = token
	g.isdebug = debug
	return nil
}

func (g *gbot) Run() (e error) {
	defer func() {
		if g.SugaredLogger != nil {
			if e != nil {
				g.Debugw("Run", "err", e)
			}
		}
	}()
	if g.status.Load() == int32(StatusRun) {
		return nil
	}
	if g.data_ch != nil {
		err := g.init(g.token, g.isdebug)
		if err != nil {
			return err
		}
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	g.ctx, g.close = context.WithCancel(context.Background())
	g.data_ch = g.api.GetUpdatesChan(u)
	go g.messageHandler(g.ctx)
	return nil
}

// 接受消息处理
func (g *gbot) messageHandler(ctx context.Context) {
	g.status.Store(int32(StatusRun))
	defer func() {
		g.status.Store(int32(StatusStop))
		g.api.StopReceivingUpdates()
	}()
	for {
		select {
		case msg := <-g.data_ch:
			if g.SugaredLogger != nil {
				g.Debugw("Message", "recive", msg)
			}
			if g.MessageHandler == nil {
				continue
			}
			g.AllMsg(g, TgMessage(msg))
			var m *Message
			//普通消息
			if msg.Message != nil && msg.Message.Chat != nil && msg.Message.From != nil {
				m = &Message{
					MessageId:        msg.Message.MessageID,
					FromUserId:       msg.Message.From.ID,
					Msg:              msg.Message.Text,
					MessageTimestamp: int64(msg.Message.Date),
					MsgType:          MessageType(msg.Message.Chat.Type),
					Base:             TgMessage(msg),
				}
				if !msg.Message.Chat.IsPrivate() {
					m.FromGroupId = msg.Message.Chat.ID
				}
			}
			//call back 消息
			if msg.CallbackQuery != nil &&
				msg.CallbackQuery.Message != nil &&
				msg.CallbackQuery.Message.Chat != nil &&
				msg.CallbackQuery.Message.From != nil {
				m = &Message{
					IsCallBack:       true,
					MessageId:        msg.CallbackQuery.Message.MessageID,
					FromUserId:       msg.CallbackQuery.From.ID,
					Msg:              msg.CallbackQuery.Data,
					MessageTimestamp: int64(msg.CallbackQuery.Message.Date),
					MsgType:          MessageType(msg.CallbackQuery.Message.Chat.Type),
					Base:             TgMessage(msg),
				}
				if !msg.CallbackQuery.Message.Chat.IsPrivate() {
					m.FromGroupId = msg.CallbackQuery.Message.Chat.ID
				}
			}
			if m.MessageId == 0 {
				continue
			}
			pass := false
			for _, t := range g.Type() {
				if t == TypeAll || string(t) == string(m.MsgType) {
					pass = true
					break
				}
			}
			if pass {
				g.ReciveMsg(g, MessageType(m.MsgType), m)
			}
		case <-ctx.Done():
			if g.SugaredLogger != nil {
				g.Debugw("MessageHandler", "status", "quit")
			}
			return
		}
	}
}

func (g *gbot) Stop() {
	if g.status.Load() == int32(StatusStop) {
		return
	}
	if g.close != nil {
		g.close()
	}
}

func (g *gbot) SendMsg(msg tgbotapi.Chattable) error {
	if g.api == nil {
		return fmt.Errorf("not init")
	}
	if msg != nil {
		_, err := g.api.Send(msg)
		return err
	}
	return nil
}

func (g *gbot) RegHandler(h MessageHandler) {
	g.MessageHandler = h
}
