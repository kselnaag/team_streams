package tg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	T "team_streams/internal/types"

	TG "github.com/go-telegram/bot"
	TGm "github.com/go-telegram/bot/models"
)

var _ T.ITG = (*Tg)(nil)

type Tg struct {
	cfg T.ICfg
	log T.ILog
	ctx context.Context
	bot *TG.Bot
}

func NewTGBot(cfg T.ICfg, log T.ILog) *Tg {
	return &Tg{
		cfg: cfg,
		log: log,
	}
}

func appRestart() {
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGHUP)
}

func (tg *Tg) isAutorized(next TG.HandlerFunc) TG.HandlerFunc {
	return func(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
		channels := make(map[string]bool, 2*len(tg.cfg.GetJsonVals()))
		for _, elem := range tg.cfg.GetJsonVals() {
			channels[elem.TgChannelID] = true
			channels[elem.TgChatID] = true
		}
		usersAutorized := make(map[int64]string, 16)
		for channel := range channels {
			admins, err := bot.GetChatAdministrators(ctx, &TG.GetChatAdministratorsParams{ChatID: channel})
			if err != nil {
				tg.log.LogDebug("TG.GetChatAdministrators(ChatID:%s) error: %s", channel, err.Error())
				continue
			}
			for _, el := range admins {
				if el.Owner != nil && !el.Owner.User.IsBot {
					usersAutorized[el.Owner.User.ID] = el.Owner.User.FirstName
					continue
				}
				if el.Administrator != nil && !el.Administrator.User.IsBot {
					usersAutorized[el.Administrator.User.ID] = el.Administrator.User.FirstName
				}
			}
		}
		tg.log.LogDebug("Message from ID:%d USERNAME:%s", update.Message.From.ID, update.Message.From.Username)
		tg.log.LogDebug("Autorized users: %+v", usersAutorized)

		for id := range usersAutorized {
			if id == update.Message.From.ID {
				next(ctx, bot, update)
			}
		}
	}
}

func (tg *Tg) infoHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /info
	var err error
	if update.Message != nil {
		cfgmsg := T.TS_APP_NAME + "=" + tg.cfg.GetEnvVal(T.TS_APP_NAME) + "\n" +
			// T.TS_APP_IP + "=" + tg.cfg.GetEnvVal(T.TS_APP_IP) + "\n" +
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
		tg.log.LogError(fmt.Errorf("%s: %w", "TG.infoHandler() error", err))
	}
}

func (tg *Tg) loglevelHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /loglevel
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		tg.log.LogDebug("loglevelHander(): %s", update.Message.Text)
		switch command[1] {
		case T.StrTrace, T.StrDebug, T.StrInfo, T.StrWarn, T.StrError, T.StrPanic, T.StrFatal, T.StrNoLog:
			tg.cfg.SetEnvVal(T.TS_LOG_LEVEL, command[1])
			appRestart()
			return
		}
	}
	tg.log.LogDebug("loglevelHander(): TRACE|DEBUG|INFO|WARN|ERROR|PANIC|FATAL|NOLOG not found")
	_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command /loglevel should use TRACE|DEBUG|INFO|WARN|ERROR|PANIC|FATAL|NOLOG flag",
	})
}

func (tg *Tg) forwardHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /autoforward
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		tg.log.LogDebug("forwardHander() %s", update.Message.Text)
		switch command[1] {
		case "ON", "OFF":
			tg.cfg.SetEnvVal(T.TS_APP_AUTOFORWARD, command[1])
			return
		}
	}
	tg.log.LogDebug("forwardHandler(): ON|OFF not found")
	_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command /autoforward should use ON|OFF flag",
	})
}

/* func (tg *Tg) getadminsHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /getadmins
	tg.log.LogInfo("getadminsHandler %s", update.Message.Text)
} */

func (tg *Tg) sendmsgHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /sendmsg
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 2 {
		tg.log.LogDebug("sendmsgHandler() %s", update.Message.Text)
		id, err := strconv.ParseInt(command[1], 10, 64)
		if err != nil {
			tg.log.LogDebug("sendmsgHandler() error %s", err.Error())
			goto LabelErr
		}
		msg := strings.Join(command[2:], " ")
		_, err = b.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: id,
			Text:   msg,
		})
		if err == nil {
			return
		}
		tg.log.LogDebug("sendmsgHandler() error %s", err.Error())
	}
LabelErr:
	tg.log.LogDebug("sendmsgHandler(): wrong flags")
	_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command /sendmsg should use UserID as 1 flag and Message as 2 flag",
	})
}

func (tg *Tg) testHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /teststream
	tg.log.LogInfo("testHandler %s", update.Message.Text)
}

func (tg *Tg) TTVUserOnlineNotify(ttvUserID string, streams [][4]string) { // TTVnotify
	tg.log.LogInfo("TTVnotify TTVUserOnline:%s %v", ttvUserID, streams)
}

func (tg *Tg) postHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /post
	tg.log.LogInfo("postHander %s", update.Message.Text)
}

func (tg *Tg) defaultHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // repost
	// tg.log.LogInfo("defaultHandler %s", update.Message.Text)
	// !!  TODO: REPOST RECOGNITION  !!
	_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command not found",
	})
}

func (tg *Tg) errorHandler(err error) {
	if errors.Is(err, TG.ErrorBadRequest) {
		err = fmt.Errorf("%s: %w", "errorHandler ErrorBadRequest 400: ", err)
	}
	if errors.Is(err, TG.ErrorUnauthorized) {
		err = fmt.Errorf("%s: %w", "errorHandler ErrorUnauthorized 401: ", err)
	}
	if errors.Is(err, TG.ErrorForbidden) {
		err = fmt.Errorf("%s: %w", "errorHandler ErrorForbidden 403: ", err)
	}
	if errors.Is(err, TG.ErrorNotFound) {
		err = fmt.Errorf("%s: %w", "errorHandler ErrorNotFound 404: ", err)
	}
	if errors.Is(err, TG.ErrorConflict) {
		err = fmt.Errorf("%s: %w", "errorHandler ErrorConflict 409: ", err)
	}
	if TG.IsTooManyRequestsError(err) {
		err = fmt.Errorf("%s: %w: retry after %d", "errorHandler TooManyRequestsError 429", err, err.(*TG.TooManyRequestsError).RetryAfter)
	}
	tg.log.LogError(err)

	// botID, _ := strconv.ParseInt(strings.Split(tg.cfg.GetEnvVal(T.TG_BOT_TOKEN), ":")[0], 10, 64)
	_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
		ChatID: "1444444741",
		Text:   "errorHandler(): " + err.Error(),
	})
}

func (tg *Tg) Start() func(err error) {
	opts := []TG.Option{
		TG.WithWorkers(len(tg.cfg.GetJsonVals())),
		TG.WithMiddlewares(tg.isAutorized),
		TG.WithErrorsHandler(tg.errorHandler),
		TG.WithDefaultHandler(tg.defaultHandler),
	}
	var err error
	tg.bot, err = TG.New(tg.cfg.GetEnvVal(T.TG_BOT_TOKEN), opts...)
	if nil != err {
		tg.log.LogError(fmt.Errorf("%s: %w", "TG_bot can not create TG_bot with error", err))
		return func(err error) {}
	}
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_INFO, TG.MatchTypeCommandStartOnly, tg.infoHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_LOGLEVEL, TG.MatchTypeCommandStartOnly, tg.loglevelHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_AUTOFORWARD, TG.MatchTypeCommandStartOnly, tg.forwardHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_TESTSTREAM, TG.MatchTypeCommandStartOnly, tg.testHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_POST, TG.MatchTypeCommandStartOnly, tg.postHandler)
	// tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_GETADMINS, TG.MatchTypeCommandStartOnly, tg.getadminsHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_SENDMSG, TG.MatchTypeCommandStartOnly, tg.sendmsgHandler)

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
