package tg

import (
	"context"
	"errors"
	"fmt"
	"os"
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

func (tg *Tg) authorized(next TG.HandlerFunc) TG.HandlerFunc {
	return func(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
		adminChannel := tg.cfg.GetJsonAdmin().TgChannelID
		usersAutorized := make(map[int64]string, 16)
		admins, err := bot.GetChatAdministrators(ctx, &TG.GetChatAdministratorsParams{ChatID: adminChannel})
		if err != nil {
			tg.log.LogDebug("TG.GetChatAdministrators(ChatID:%s) error: %s", adminChannel, err.Error())
			return
		}
		for _, el := range admins {
			if el.Owner != nil && !el.Owner.User.IsBot {
				usersAutorized[el.Owner.User.ID] = el.Owner.User.FirstName
			}
			if el.Administrator != nil && !el.Administrator.User.IsBot {
				usersAutorized[el.Administrator.User.ID] = el.Administrator.User.FirstName
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

func (tg *Tg) infoHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /info
	var err error
	if update.Message != nil {
		cfgmsg := T.TS_APP_NAME + "=" + tg.cfg.GetEnvVal(T.TS_APP_NAME) + "\n" +
			// T.TS_APP_IP + "=" + tg.cfg.GetEnvVal(T.TS_APP_IP) + "\n" +
			T.TS_LOG_LEVEL + "=" + tg.cfg.GetEnvVal(T.TS_LOG_LEVEL) + "\n" +
			T.TS_APP_AUTOFORWARD + "=" + tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) + "\n"
		usermsg := make([]string, 0, len(tg.cfg.GetJsonUsers())+1)
		usermsg = append(usermsg, "Users:")
		for _, el := range tg.cfg.GetJsonUsers() {
			usermsg = append(usermsg, el.Nickname)
		}
		_, err = bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   cfgmsg + strings.Join(usermsg, "  "),
		})
	}
	if err != nil {
		tg.log.LogError(fmt.Errorf("%s: %w", "TG.infoHandler() error", err))
	}
}

func (tg *Tg) loglevelHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /loglevel
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
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command /loglevel should use TRACE|DEBUG|INFO|WARN|ERROR|PANIC|FATAL|NOLOG as 1st flag",
	})
}

func (tg *Tg) forwardHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /autoforward
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		tg.log.LogDebug("forwardHander() %s", update.Message.Text)
		switch command[1] {
		case "DEBUG", "OFF", "ON":
			tg.cfg.SetEnvVal(T.TS_APP_AUTOFORWARD, command[1])
			return
		}
	}
	tg.log.LogDebug("forwardHandler(): DEBUG|OFF|ON not found")
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command /autoforward should use DEBUG|OFF|ON as 1st flag",
	})
}

func (tg *Tg) getadminsHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /getadmins
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		tg.log.LogDebug("getadminsHandler() %s", update.Message.Text)
		admins, err := bot.GetChatAdministrators(ctx, &TG.GetChatAdministratorsParams{ChatID: command[1]})
		if err != nil {
			tg.log.LogDebug("TG.GetChatAdministrators(ChatID:%s) error: %s", command[1], err.Error())
			return
		}
		parsedAdmins := make([][2]string, 0, 16)
		for _, el := range admins {
			if el.Owner != nil && !el.Owner.User.IsBot {
				parsedAdmins = append(parsedAdmins, [2]string{fmt.Sprintf("%d", el.Owner.User.ID), el.Owner.User.FirstName})
			}
			if el.Administrator != nil && !el.Administrator.User.IsBot {
				parsedAdmins = append(parsedAdmins, [2]string{fmt.Sprintf("%d", el.Administrator.User.ID), el.Administrator.User.FirstName})
			}
		}
		_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("%v", parsedAdmins),
		})
		return
	}
	tg.log.LogDebug("getadminsdHandler(): chatID or channelID not found")
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command /getadmins should use chatID or channelID as 1st flag",
	})
}

func (tg *Tg) sendmsgHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /sendmsg
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 2 {
		tg.log.LogDebug("sendmsgHandler() %s", update.Message.Text)
		msg := strings.Join(command[2:], " ")
		_, err := b.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: command[1],
			Text:   msg,
		})
		if err != nil {
			tg.log.LogDebug("sendmsgHandler() error %s", err.Error())
			_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("sendmsgHandler() error %s", err.Error()),
			})
		}
		return
	}
	tg.log.LogDebug("sendmsgHandler(): wrong flags")
	_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command /sendmsg should use UserID as 1 flag and Message as 2 flag",
	})
}

func (tg *Tg) testHandler(ctx context.Context, b *TG.Bot, update *TGm.Update) { // /teststream
	tg.log.LogDebug("testHandler %s", update.Message.Text)
	adminUser := tg.cfg.GetJsonAdmin()
	adminChanID := adminUser.TgChannelID
	adminChatID := adminUser.TgChatID

	msg := "(Тестовое уведомление)\n" +
		"Возрадуйтесь братья и сестры!\n" + adminUser.Longname + " соизволил запустить стрим!\n\n" +
		"kselnaag" + " || " + "Software and Game Development\n" +
		"Пишем Cтойло, смотрим стрим\n\n" +
		"https://www.twitch.tv/" + "kselnaag"
	_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: adminChanID,
		Text:   msg,
	})

	switch tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) {
	case "ON":
		for _, el := range tg.cfg.GetJsonUsers() {
			_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: el.TgChannelID,
				Text:   msg,
			})
		}
	case "OFF":
		_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: adminChatID,
			Text:   msg,
		})
	}
}

func (tg *Tg) TTVUserOnlineNotify(ttvUserID string, ttvStreams [][4]string) { // TTVnotify
	tg.log.LogInfo("TTVnotify TTVUserOnline:%s %v", ttvUserID, ttvStreams)
	var (
		tgUser  T.User
		ttvUser [4]string // [4]string{elem.UserID, elem.UserLogin, elem.GameName, elem.Title}
	)
	for _, el := range tg.cfg.GetJsonUsers() {
		if el.TtvUserID == ttvUserID {
			tgUser = el
			break
		}
	}
	for _, elem := range ttvStreams {
		if elem[0] == ttvUserID {
			ttvUser = elem
			break
		}
	}
	msg := "Возрадуйтесь братья и сестры!\n" + tgUser.Longname + " соизволила запустить стрим!\n\n" +
		ttvUser[1] + " || " + ttvUser[2] + "\n" +
		ttvUser[3] + "\n\n" + "https://www.twitch.tv/" + ttvUser[1]
	_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
		ChatID: tg.cfg.GetJsonAdmin().TgChannelID,
		Text:   msg,
	})

	switch tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) {
	case "ON":
		for _, el := range tg.cfg.GetJsonUsers() {
			_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
				ChatID: el.TgChannelID,
				Text:   msg,
			})
		}
	case "OFF":
		_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
			ChatID: tgUser.TgChannelID,
			Text:   msg,
		})
	}
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
		err = fmt.Errorf("%s: %w", "errorHandler() ErrorBadRequest 400: ", err)
	}
	if errors.Is(err, TG.ErrorUnauthorized) {
		err = fmt.Errorf("%s: %w", "errorHandler() ErrorUnauthorized 401: ", err)
	}
	if errors.Is(err, TG.ErrorForbidden) {
		err = fmt.Errorf("%s: %w", "errorHandler() ErrorForbidden 403: ", err)
	}
	if errors.Is(err, TG.ErrorNotFound) {
		err = fmt.Errorf("%s: %w", "errorHandler() ErrorNotFound 404: ", err)
	}
	if errors.Is(err, TG.ErrorConflict) {
		err = fmt.Errorf("%s: %w", "errorHandler() ErrorConflict 409: ", err)
	}
	if TG.IsTooManyRequestsError(err) {
		err = fmt.Errorf("%s: %w: retry after %d", "errorHandler() TooManyRequestsError 429", err, err.(*TG.TooManyRequestsError).RetryAfter)
	}
	tg.log.LogError(err)
	_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
		ChatID: "1444444741",
		Text:   "errorHandler(): " + err.Error(),
	})
}

func (tg *Tg) Start() func(err error) {
	opts := []TG.Option{
		TG.WithWorkers(len(tg.cfg.GetJsonUsers())),
		TG.WithMiddlewares(tg.authorized),
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
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, T.COMMAND_GETADMINS, TG.MatchTypeCommandStartOnly, tg.getadminsHandler)
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
