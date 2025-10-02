package tg

import (
	"context"
	"fmt"
	"strings"
	T "team_streams/internal/types"

	TG "github.com/go-telegram/bot"
	TGm "github.com/go-telegram/bot/models"
)

func (tg *Tg) authorized(next TG.HandlerFunc) TG.HandlerFunc {
	return func(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
		adminChannel := tg.cfg.GetJsonAdmin().TgChannelID
		admins, err := bot.GetChatAdministrators(ctx, &TG.GetChatAdministratorsParams{ChatID: adminChannel})
		if err != nil {
			err = fmt.Errorf("TG.GetChatAdministrators(ChatID:%s) error: %w", adminChannel, err)
			tg.log.LogError(err)
			_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   err.Error(),
			})
			return
		}
		usersAutorized := make(map[int64]string, len(admins))
		for _, el := range admins {
			if el.Owner != nil && !el.Owner.User.IsBot {
				usersAutorized[el.Owner.User.ID] = el.Owner.User.FirstName + "*"
			}
			if el.Administrator != nil && !el.Administrator.User.IsBot {
				usersAutorized[el.Administrator.User.ID] = el.Administrator.User.FirstName
			}
		}
		tg.log.LogDebug("Message from ID:%d USERNAME:%s MSG:%s", update.Message.From.ID, update.Message.From.Username, update.Message.Text)
		tg.log.LogDebug("Autorized users: %+v", usersAutorized)
		for id := range usersAutorized {
			if id == update.Message.From.ID {
				next(ctx, bot, update)
			}
		}
	}
}

func (tg *Tg) infoHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /info
	tg.log.LogDebug("infolHander(): %s", update.Message.Text)
	cfgmsg := T.TS_APP_NAME + "=" + tg.cfg.GetEnvVal(T.TS_APP_NAME) + "\n" +
		T.TS_LOG_LEVEL + "=" + tg.cfg.GetEnvVal(T.TS_LOG_LEVEL) + "\n" +
		T.TS_APP_AUTOFORWARD + "=" + tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) + "\n"
	admins := make([]string, 0, len(tg.cfg.GetJsonUsers())+1)
	admins = append(admins, "Admins: "+tg.cfg.GetJsonAdmin().Nickname+"*")
	for _, el := range tg.cfg.GetJsonUsers() {
		admins = append(admins, el.Nickname)
	}
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   cfgmsg + strings.Join(admins, " "),
	})
}

func (tg *Tg) loglevelHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /loglevel
	tg.log.LogDebug("loglevelHander(): %s", update.Message.Text)
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		switch command[1] {
		case T.StrTrace, T.StrDebug, T.StrInfo, T.StrWarn, T.StrError, T.StrPanic, T.StrFatal, T.StrNoLog:
			tg.cfg.SetEnvVal(T.TS_LOG_LEVEL, command[1])
			tg.appRestart()
			return
		}
	}
	str := T.StrTrace + "|" + T.StrDebug + "|" + T.StrInfo + "|" + T.StrWarn + "|" + T.StrError + "|" + T.StrPanic + "|" + T.StrFatal + "|" + T.StrNoLog
	tg.log.LogDebug("loglevelHander():%s not found", str)
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("command /loglevel should use %s as 1st flag", str),
	})
}

func (tg *Tg) forwardHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /autoforward
	tg.log.LogDebug("forwardHandler(): %s", update.Message.Text)
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		switch command[1] {
		case T.AFORW_DEBUG, T.AFORW_OFF, T.AFORW_ON:
			tg.cfg.SetEnvVal(T.TS_APP_AUTOFORWARD, command[1])
		}
		return
	}
	str := T.AFORW_DEBUG + "|" + T.AFORW_OFF + "|" + T.AFORW_ON
	tg.log.LogDebug("forwardHandler(): %s not found", str)
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("command /autoforward should use %s as 1st flag", str),
	})
}

func (tg *Tg) getadminsHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /getadmins
	tg.log.LogDebug("getadminsHandler() %s", update.Message.Text)
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		admins, err := bot.GetChatAdministrators(ctx, &TG.GetChatAdministratorsParams{ChatID: command[1]})
		if err != nil {
			err = fmt.Errorf("TG.GetChatAdministrators(ChatID: %s) error: %w", command[1], err)
			tg.log.LogError(err)
			_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   err.Error(),
			})
			return
		}
		parsedAdmins := make([][2]string, 0, len(admins))
		for _, el := range admins {
			if el.Owner != nil && !el.Owner.User.IsBot {
				parsedAdmins = append(parsedAdmins, [2]string{fmt.Sprintf("%d", el.Owner.User.ID), el.Owner.User.FirstName + "*"})
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
	tg.log.LogDebug("sendmsgHandler() %s", update.Message.Text)
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 2 {
		msg := strings.Join(command[2:], " ")
		_, err := b.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: command[1],
			Text:   msg,
		})
		if err != nil {
			tg.log.LogError(fmt.Errorf("sendmsgHandler() error: %w", err))
			_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("sendmsgHandler() error: %s", err.Error()),
			})
		}
		return
	}
	str := "UserID as 1 flag and Message as 2 flag"
	tg.log.LogDebug("sendmsgHandler(): %s not found", str)
	_, _ = b.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("command /sendmsg should use %s", str),
	})
}

func (tg *Tg) testHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /teststream
	tg.log.LogDebug("testHandler() %s", update.Message.Text)
	admin := tg.cfg.GetJsonAdmin()
	msg := "(Тестовое уведомление)\n" +
		"Возрадуйтесь братья и сестры!\n" + admin.Longname + " соизволил запустить стрим!\n\n" +
		admin.Nickname + " || " + "Game title here\n" +
		"Stream title here\n\n" +
		"https://www.twitch.tv/" + admin.Nickname
	_, errDEBUG := bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: admin.TgChannelID,
		Text:   msg,
	})
	if errDEBUG != nil {
		tg.log.LogDebug("testHandler() error: %s", errDEBUG.Error())
		_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("command /teststream error: %s", errDEBUG.Error()),
		})
	}
	switch tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) {
	case T.AFORW_ON:
		tg.sendToAllUsers(ctx, bot, msg)
	case T.AFORW_OFF:
		_, errOFF := bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: admin.TgChatID, // for testing
			Text:   msg,
		})
		if errOFF != nil {
			tg.log.LogDebug("testHandler() OFF error: %s: ChanID[%s]", errOFF.Error(), admin.TgChatID)
			_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("command /teststream error: %s", errOFF.Error()),
			})
		}
	}
}

func (tg *Tg) sendToAllUsers(ctx context.Context, bot *TG.Bot, msg string) {
	for _, el := range tg.cfg.GetJsonUsers() {
		go func() {
			_, errON := bot.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: el.TgChannelID,
				Text:   msg,
			})
			if errON != nil {
				tg.log.LogDebug("sendToAllUsers() AF_ON: ChanID[%s] error: %s", el.TgChannelID, errON.Error())
			}
		}()
	}
}

func (tg *Tg) TTVUserOnlineNotify(ttvUserID string, ttvStreams [][4]string) { // TTVnotify
	tg.log.LogDebug("TTVnotify() Online:%s %v", ttvUserID, ttvStreams)
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
	_, errDEBUG := tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
		ChatID: tg.cfg.GetJsonAdmin().TgChannelID,
		Text:   msg,
	})
	if errDEBUG != nil {
		tg.log.LogDebug("TTVnotify()_sendMessage() error: ChanID[%s]: %s", tg.cfg.GetJsonAdmin().TgChannelID, errDEBUG.Error())
		_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
			ChatID: tg.cfg.GetJsonAdmin().TgChatID,
			Text:   fmt.Sprintf("TTVUserOnlineNotify DEBUG() error: %s", errDEBUG.Error()),
		})
	}
	switch tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) {
	case "ON":
		tg.sendToAllUsers(tg.ctx, tg.bot, msg)
	case "OFF":
		_, errOFF := tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
			ChatID: tgUser.TgChannelID,
			Text:   msg,
		})
		if errOFF != nil {
			tg.log.LogDebug("TTVnotify() error: ChanID[%s]: %s", tgUser.TgChannelID, errOFF.Error())
			_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
				ChatID: tg.cfg.GetJsonAdmin().TgChatID,
				Text:   fmt.Sprintf("TTVUserOnlineNotify OFF() error: %s", errOFF.Error()),
			})
		}
	}
}

func (tg *Tg) postHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /post
	tg.log.LogDebug("postHander() %s", update.Message.Text)
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		msg := strings.Join(command[1:], " ")
		_, errDEBUG := bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: tg.cfg.GetJsonAdmin().TgChannelID,
			Text:   msg,
		})
		if errDEBUG != nil {
			tg.log.LogDebug("postHandler() error: %s: ChanID[%s]", errDEBUG.Error(), tg.cfg.GetJsonAdmin().TgChannelID)
			_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("command /post error: %s", errDEBUG.Error()),
			})
		}
		tg.sendToAllUsers(ctx, bot, msg)
		return
	}
	tg.log.LogDebug("postHander(): should use Message as 1 flag")
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command /post should use Message as 1 flag",
	})
}

func (tg *Tg) defaultHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // repost
	tg.log.LogDebug("defaultHander(): %s", update.Message.Text)
	if update.Message.ForwardOrigin != nil {
		_, err := bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: tg.cfg.GetJsonAdmin().TgChannelID,
			Text:   update.Message.Text,
		})
		if err != nil {
			tg.log.LogDebug("defaultHandler() error: ChanID[%s]: %s", tg.cfg.GetJsonAdmin().TgChannelID, err.Error())
			_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("forward error: %s", err.Error()),
			})
		}
		tg.sendToAllUsers(ctx, bot, update.Message.Text)
		return
	}
	tg.log.LogDebug("defaultHander(): command not found")
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command not found",
	})
}
