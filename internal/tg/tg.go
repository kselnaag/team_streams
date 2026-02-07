package tg

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	T "team_streams/internal/types"
	"team_streams/pic"

	TG "github.com/go-telegram/bot"
)

var _ T.ITG = (*Tg)(nil)

type delMsg struct {
	ChanID string
	MsgID  int
}

type Tg struct {
	cfg       T.ICfg
	log       T.ILog
	ctx       context.Context
	bot       *TG.Bot
	fs        embed.FS
	mu        sync.Mutex
	msgsToDel [][]delMsg
}

func NewTGBot(cfg T.ICfg, log T.ILog) *Tg {
	msgsToDel := make([][]delMsg, len(cfg.GetJsonUsers()))
	for i := range msgsToDel {
		msgsToDel[i] = make([]delMsg, 0, 4)
	}
	return &Tg{
		cfg:       cfg,
		log:       log,
		fs:        pic.StaticFS,
		msgsToDel: msgsToDel,
	}
}

func (tg *Tg) appRestart() {
	tg.log.LogDebug("appRestart() called")
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGHUP)
}

func (tg *Tg) errorHandler(err error) {
	if errors.Is(err, TG.ErrorBadRequest) {
		err = fmt.Errorf("%s: %w", "ErrorBadRequest 400: ", err)
	}
	if errors.Is(err, TG.ErrorUnauthorized) {
		err = fmt.Errorf("%s: %w", "ErrorUnauthorized 401: ", err)
	}
	if errors.Is(err, TG.ErrorForbidden) {
		err = fmt.Errorf("%s: %w", "ErrorForbidden 403: ", err)
	}
	if errors.Is(err, TG.ErrorNotFound) {
		err = fmt.Errorf("%s: %w", "ErrorNotFound 404: ", err)
	}
	if errors.Is(err, TG.ErrorConflict) {
		err = fmt.Errorf("%s: %w", "ErrorConflict 409: ", err)
	}
	if TG.IsTooManyRequestsError(err) {
		err = fmt.Errorf("TooManyRequestsError 429: retry after %d: %w", err.(*TG.TooManyRequestsError).RetryAfter, err)
	}
	tg.log.LogError(err)
	_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
		ChatID: tg.cfg.GetJsonAdmin().TgUserID,
		Text:   "TG errorHandler(): " + err.Error(),
	})
}

func (tg *Tg) Start() func(err error) {
	opts := []TG.Option{
		TG.WithMiddlewares(tg.notifyAutoforwardDelete),
		TG.WithMiddlewares(tg.authorized),
		TG.WithDefaultHandler(tg.defaultHandler),
		TG.WithErrorsHandler(tg.errorHandler),
	}
	var err error
	tg.bot, err = TG.New(tg.cfg.GetEnvVal(T.TG_BOT_TOKEN), opts...)
	if nil != err {
		tg.log.LogError(fmt.Errorf("TG_bot can not create TG_bot with error: %w", err))
		return func(err error) {}
	}
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_INFO, TG.MatchTypeCommandStartOnly, tg.infoHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_LOGLEVEL, TG.MatchTypeCommandStartOnly, tg.loglevelHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_AUTOFORWARD, TG.MatchTypeCommandStartOnly, tg.forwardHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_AUTODEL, TG.MatchTypeCommandStartOnly, tg.delHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_TESTSTREAM, TG.MatchTypeCommandStartOnly, tg.testHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_POST, TG.MatchTypeCommandStartOnly, tg.postHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_DELALL, TG.MatchTypeCommandStartOnly, tg.delallHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_GETADMINS, TG.MatchTypeCommandStartOnly, tg.getadminsHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_SENDMSG, TG.MatchTypeCommandStartOnly, tg.sendmsgHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_HELP, TG.MatchTypeCommandStartOnly, tg.helpHandler)

	var ctxCancelTGbot context.CancelFunc
	tg.ctx, ctxCancelTGbot = context.WithCancel(context.Background())
	go tg.bot.Start(tg.ctx)
	tg.log.LogInfo("TG_bot started")
	return func(err error) { // TgStop
		ctxCancelTGbot()
		if err != nil {
			tg.log.LogError(fmt.Errorf("%s: %w", "TG_bot stoped with error", err))
		} else {
			tg.log.LogInfo("TG_bot stoped")
		}
	}
}
