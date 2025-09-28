package tg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	T "team_streams/internal/types"

	TG "github.com/go-telegram/bot"
	TGm "github.com/go-telegram/bot/models"
)

var _ T.ITG = (*Tg)(nil)

type Tg struct {
	cfg   T.ICfg
	log   T.ILog
	tgbot *TG.Bot
}

func NewTGBot(cfg T.ICfg, log T.ILog) *Tg {
	return &Tg{
		cfg: cfg,
		log: log,
	}
}

func (tg *Tg) infoHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /info
	// tg.log.LogInfo("infoHandler %s", update.Message.Text)
	var err error
	if update.Message != nil {
		cfgmsg := T.TS_APP_NAME + "=" + tg.cfg.GetEnvVal(T.TS_APP_NAME) + "\n" +
			T.TS_APP_IP + "=" + tg.cfg.GetEnvVal(T.TS_APP_IP) + "\n" +
			T.TS_LOG_LEVEL + "=" + tg.cfg.GetEnvVal(T.TS_LOG_LEVEL) + "\n" +
			T.TS_APP_AUTOFORWARD + "=" + tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) + "\n"
		usermsg := make([]string, 0, len(tg.cfg.GetJsonVals())+1)
		usermsg = append(usermsg, "Users:  ")
		for _, el := range tg.cfg.GetJsonVals() {
			var dop string
			if el.TgUserID != "" {
				dop = "* "
			} else {
				dop = "  "
			}
			usermsg = append(usermsg, el.Nickname+dop)
		}
		_, err = b.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   cfgmsg + strings.Join(usermsg, ""),
		})
	}
	if err != nil {
		tg.log.LogError(fmt.Errorf("%s: %w", "TG.infoHandler()", err))
	}
}

func (tg *Tg) loglevelHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /loglevel
	tg.log.LogInfo("loglevelHander %s", update.Message.Text)
}

func (tg *Tg) forwardHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /autoforward
	tg.log.LogInfo("autoforwardHander %s", update.Message.Text)
}

func (tg *Tg) testHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /teststream
	tg.log.LogInfo("testHandler %s", update.Message.Text)
}

func (tg *Tg) postHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /post
	tg.log.LogInfo("postHander %s", update.Message.Text)
}

func (tg *Tg) getadminsHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /getadmins
	tg.log.LogInfo("getadminsHandler %s", update.Message.Text)
}

func (tg *Tg) sendmsgHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /sendmsg
	tg.log.LogInfo("sendmsgHandler %s", update.Message.Text)
}

func (tg *Tg) TTVUserOnlineNotify(ttvUserID string, streams [][4]string) { // info from TTV
	tg.log.LogInfo("TTVnotify TTVUserOnline:%s %v", ttvUserID, streams)
}

func (tg *Tg) defaultHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) {
	tg.log.LogInfo("defaultHandler %s", update.Message.Text)
	/* 	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Say message with `/foo` anywhere or with `/bar` at start of the message",
		ParseMode: models.ParseModeMarkdown,
	}) */
}

func (tg *Tg) errorHandler(err error) {
	if errors.Is(err, TG.ErrorBadRequest) {
		tg.log.LogError(fmt.Errorf("%s: %w", "errorHandler ErrorBadRequest 400: ", err))
	}
	if errors.Is(err, TG.ErrorUnauthorized) {
		tg.log.LogError(fmt.Errorf("%s: %w", "errorHandler ErrorUnauthorized 401: ", err))
	}
	if errors.Is(err, TG.ErrorForbidden) {
		tg.log.LogError(fmt.Errorf("%s: %w", "errorHandler ErrorForbidden 403: ", err))
	}
	if errors.Is(err, TG.ErrorNotFound) {
		tg.log.LogError(fmt.Errorf("%s: %w", "errorHandler ErrorNotFound 404: ", err))
	}
	if errors.Is(err, TG.ErrorConflict) {
		tg.log.LogError(fmt.Errorf("%s: %w", "errorHandler ErrorConflict 409: ", err))
	}
	if TG.IsTooManyRequestsError(err) {
		tg.log.LogError(fmt.Errorf("%s: %s: retry after %d: %w", "errorHandler TooManyRequestsError 429", err.Error(), err.(*TG.TooManyRequestsError).RetryAfter, err))
	}
}

func (tg *Tg) Start() func(err error) {
	opts := []TG.Option{
		TG.WithWorkers(len(tg.cfg.GetJsonVals())),
		TG.WithErrorsHandler(tg.errorHandler),
		TG.WithDefaultHandler(tg.defaultHandler),
	}
	var err error
	tg.tgbot, err = TG.New(tg.cfg.GetEnvVal(T.TG_BOT_TOKEN), opts...)
	if nil != err {
		tg.log.LogError(fmt.Errorf("%s: %w", "TG_bot can not create TG_bot with error", err))
		return func(err error) {}
	}
	tg.tgbot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_INFO, TG.MatchTypeCommandStartOnly, tg.infoHandler)
	tg.tgbot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_LOGLEVEL, TG.MatchTypeCommandStartOnly, tg.loglevelHandler)
	tg.tgbot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_AUTOFORWARD, TG.MatchTypeCommandStartOnly, tg.forwardHandler)
	tg.tgbot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_TESTSTREAM, TG.MatchTypeCommandStartOnly, tg.testHandler)
	tg.tgbot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_POST, TG.MatchTypeCommandStartOnly, tg.postHandler)
	tg.tgbot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_GETADMINS, TG.MatchTypeCommandStartOnly, tg.getadminsHandler)
	tg.tgbot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_SENDMSG, TG.MatchTypeCommandStartOnly, tg.sendmsgHandler)

	ctxTGbot, ctxCancelTGbot := context.WithCancel(context.Background())
	go tg.tgbot.Start(ctxTGbot)
	tg.log.LogInfo("TG_bot started")
	return func(err error) { // TgStop
		tg.tgbot.Close(ctxTGbot)
		ctxCancelTGbot()
		if err != nil {
			tg.log.LogError(fmt.Errorf("%s: %w", "TG_bot stoped with error", err))
		} else {
			tg.log.LogInfo("TG_bot stoped")
		}
	}
}
