package tg

import (
	"context"
	"errors"
	"fmt"
	T "team_streams/internal/types"

	TG "github.com/go-telegram/bot"
	TGm "github.com/go-telegram/bot/models"
)

var _ T.ITG = (*Tg)(nil)

type Tg struct {
	cfg   T.ICfg
	log   T.ILog
	gtbot *TG.Bot
}

func NewTGBot(cfg T.ICfg, log T.ILog) *Tg {
	opts := []TG.Option{
		TG.WithWorkers(5),
		TG.WithErrorsHandler(errorHandler),
		TG.WithDefaultHandler(defaultHandler),
	}
	gtbot, err := TG.New(cfg.GetEnvVal(T.TG_BOT_TOKEN), opts...)
	if nil != err {
		log.LogError(fmt.Errorf("%s: %w", "TG_bot can not create TG_bot with error", err))
	}
	// bot.RegisterHandler(bot.HandlerTypeMessageText, "foo", bot.MatchTypeCommand, fooHandler)
	// bot.RegisterHandler(bot.HandlerTypeMessageText, "bar", bot.MatchTypeCommandStartOnly, barHandler)
	return &Tg{
		cfg:   cfg,
		log:   log,
		gtbot: gtbot,
	}
}

func errorHandler(err error) {
	if errors.Is(err, TG.ErrorBadRequest) { // Handle the ErrorBadRequest (400) case here
	}
	if errors.Is(err, TG.ErrorUnauthorized) { // Handle the ErrorUnauthorized (401) case here
	}
	if errors.Is(err, TG.ErrorForbidden) { // Handle the ErrorForbidden (403) case here
	}
	if errors.Is(err, TG.ErrorNotFound) { // Handle the ErrorNotFound (404) case here
	}
	if errors.Is(err, TG.ErrorConflict) { // Handle the ErrorConflict (409) case here
	}
	if TG.IsTooManyRequestsError(err) { // Handle the TooManyRequestsError (429) case here
		fmt.Println("Received TooManyRequestsError with retry_after:", err.(*TG.TooManyRequestsError).RetryAfter)
	}
}

func defaultHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) {
	/* 	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Say message with `/foo` anywhere or with `/bar` at start of the message",
		ParseMode: models.ParseModeMarkdown,
	}) */
}

/*
	appinfo + usersinfo
	loglevel (TRACE, DEBUG, INFO, WARN, ERROR, PANIC, FATAL, NOLOG(default if empty or mess))
	ttvautoforward (ON/OFF)
	ttvteststream
	botprivate_auotpost
*/

func (tg *Tg) TTVUserOnlineNotify(ttvUserID string, streams [][4]string) { // info from TTV

}

func (tg *Tg) Start() func(err error) { // tg.cfg.GetJsonVals 	// tg.gtbot.GetChatAdministrators()
	ctxTGbot, ctxCancelTGbot := context.WithCancel(context.Background())
	tg.gtbot.Start(ctxTGbot)
	tg.log.LogInfo("TG_bot started")
	return func(err error) { // TgStop
		tg.gtbot.Close(ctxTGbot)
		ctxCancelTGbot()
		if err != nil {
			tg.log.LogError(fmt.Errorf("%s: %w", "TG_bot stoped with error", err))
		} else {
			tg.log.LogInfo("TG_bot stoped")
		}
	}
}
