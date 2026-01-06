package tg

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	T "team_streams/internal/types"
	"time"

	TG "github.com/go-telegram/bot"
	TGm "github.com/go-telegram/bot/models"
)

func (tg *Tg) authorized(next TG.HandlerFunc) TG.HandlerFunc {
	return func(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
		usersAutorized := tg.getChatAdmins(tg.cfg.GetJsonAdmin().TgChannelID, update)
		for id := range *usersAutorized {
			if (update.Message != nil) && (id == update.Message.From.ID) {
				next(ctx, bot, update)
			}
		}
	}
}

func (tg *Tg) getChatAdmins(channel string, update *TGm.Update) *map[int64]string {
	admins, err := tg.bot.GetChatAdministrators(tg.ctx, &TG.GetChatAdministratorsParams{ChatID: channel})
	if err != nil {
		err = fmt.Errorf("TG.getChatAdmins() ChatID:%s error: %w", channel, err)
		tg.log.LogError(err)
		_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
		})
		return &map[int64]string{}
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
	return &usersAutorized
}

func (tg *Tg) getChanInfo(ctx context.Context, bot *TG.Bot) string {
	ids := make([]string, 0, 8)
	titles := make([]string, 0, 8)
	channels := make([]string, 0, 8)
	ids = append(ids, tg.cfg.GetJsonAdmin().TgChannelID, tg.cfg.GetJsonAdmin().TgChatID)
	for _, el := range tg.cfg.GetJsonUsers() {
		ids = append(ids, el.TgChannelID, el.TgChatID)
	}
	for _, id := range ids {
		chat, err := bot.GetChat(ctx, &TG.GetChatParams{
			ChatID: id,
		})
		if err != nil {
			titles = append(titles, "ERROR")

		} else {
			titles = append(titles, chat.Title)
		}
	}
	for idx := range ids {
		channels = append(channels, "["+ids[idx]+"] "+titles[idx]+"\n")
	}
	return strings.Join(channels, "")
}

func (tg *Tg) infoHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /info
	tg.log.LogDebug("infolHander(): %s", update.Message.Text)
	cfgmsg := T.TS_APP_NAME + "=" + tg.cfg.GetEnvVal(T.TS_APP_NAME) + "\n" +
		T.TS_APP_IP + "=" + tg.cfg.GetEnvVal(T.TS_APP_IP) + "\n" +
		T.TS_APP_AUTOFORWARD + "=" + tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) + "\n" +
		T.TS_APP_AUTODEL + "=" + tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) + "\n" +
		T.TS_LOG_LEVEL + "=" + tg.cfg.GetEnvVal(T.TS_LOG_LEVEL) + "\n"
	usersAutorized := tg.getChatAdmins(tg.cfg.GetJsonAdmin().TgChannelID, update)
	admins := "Admins:\n" + fmt.Sprintf("%v", usersAutorized) + "\n"
	channels := "Channels:\n" + tg.getChanInfo(ctx, bot) + "\n"
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   cfgmsg + admins + channels,
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
			return
		}
	}
	str := T.AFORW_DEBUG + "|" + T.AFORW_OFF + "|" + T.AFORW_ON
	tg.log.LogDebug("forwardHandler(): %s not found", str)
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("command /autoforward should use %s as 1st flag", str),
	})
}

func (tg *Tg) delHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /autodel
	tg.log.LogDebug("forwardHandler(): %s", update.Message.Text)
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		switch command[1] {
		case T.ADEL_OFF, T.AFORW_ON:
			tg.cfg.SetEnvVal(T.TS_APP_AUTODEL, command[1])
			return
		}
	}
	str := T.ADEL_OFF + "|" + T.ADEL_ON
	tg.log.LogDebug("delHandler(): %s not found", str)
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("command /autodel should use %s as 1st flag", str),
	})
}

func (tg *Tg) getadminsHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /getadmins
	tg.log.LogDebug("getadminsHandler() %s", update.Message.Text)
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		usersAutorized := tg.getChatAdmins(command[1], update)
		_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("%v", usersAutorized),
		})
		return
	}
	tg.log.LogDebug("getadminsdHandler(): chatID or channelID not found")
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command /getadmins should use chatID or channelID as 1st flag",
	})
}

func (tg *Tg) sendmsgHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /sendmsg
	tg.log.LogDebug("sendmsgHandler() %s", update.Message.Text)
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 2 {
		msg := strings.Join(command[2:], " ")
		_, err := bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: command[1],
			Text:   msg,
		})
		if err != nil {
			tg.log.LogError(fmt.Errorf("sendmsgHandler() error: %w", err))
			_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("sendmsgHandler() error: %s", err.Error()),
			})
		}
		return
	}
	str := "UserID or channelID as 1 flag and MessageText as 2 flag"
	tg.log.LogDebug("sendmsgHandler(): %s not found", str)
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("command /sendmsg should use %s", str),
	})
}

func (tg *Tg) testHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /teststream
	tg.log.LogDebug("testHandler() %s", update.Message.Text)
	admin := tg.cfg.GetJsonAdmin()
	msg := "(Тестовое уведомление)\n" +
		"Возрадуйтесь братья и сестры!\n" + admin.Longname + " соизволил запустить стрим !\n\n" +
		admin.Nickname + "  |  " + "Software and game development\n" +
		"Пишем бота, смотрим стрим!\n\n" +
		"https://www.twitch.tv/" + admin.Nickname

	fileData, _ := tg.fs.ReadFile("data/" + admin.Nickname + "_pic.jpg")
	sentMsg, errDEBUG := bot.SendPhoto(ctx, &TG.SendPhotoParams{
		ChatID:  admin.TgChannelID,
		Photo:   &TGm.InputFileUpload{Filename: admin.Nickname + "_pic.jpg", Data: bytes.NewReader(fileData)},
		Caption: msg,
	})
	if errDEBUG != nil {
		tg.log.LogDebug("testHandler() DEBUG error: %s", errDEBUG.Error())
		_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("command /teststream error: %s", errDEBUG.Error()),
		})
	}
	switch tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) {
	case T.AFORW_ON:
		for _, el := range tg.cfg.GetJsonUsers() {
			go func() {
				fwdMsg, errON := bot.ForwardMessage(ctx, &TG.ForwardMessageParams{
					ChatID:     el.TgChannelID,
					FromChatID: admin.TgChannelID,
					MessageID:  sentMsg.ID,
				})
				if errON != nil {
					tg.log.LogDebug("testHandler() ON error: %s: ChanID[%s]", errON.Error(), el.TgChannelID)
					_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("command /teststream error: %s: ChanID[%s]", errON.Error(), el.TgChannelID),
					})
				}
				if tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) == T.ADEL_ON {
					time.Sleep(10 * time.Second)
					_, _ = tg.bot.DeleteMessage(tg.ctx, &TG.DeleteMessageParams{
						ChatID:    el.TgChannelID,
						MessageID: fwdMsg.ID,
					})
				}
			}()
		}
		fallthrough
	case T.AFORW_OFF:
		fwdMsg, errOFF := bot.ForwardMessage(ctx, &TG.ForwardMessageParams{
			ChatID:     admin.TgChatID, // for testing only
			FromChatID: admin.TgChannelID,
			MessageID:  sentMsg.ID,
		})
		if errOFF != nil {
			tg.log.LogDebug("testHandler() OFF error: %s: ChanID[%s]", errOFF.Error(), admin.TgChatID)
			_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("command /teststream error: %s", errOFF.Error()),
			})
		}
		if tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) == T.ADEL_ON {
			time.Sleep(10 * time.Second)
			_, _ = tg.bot.DeleteMessage(tg.ctx, &TG.DeleteMessageParams{
				ChatID:    admin.TgChatID, // for testing only
				MessageID: fwdMsg.ID,
			})
		}
	}
}

func (tg *Tg) helpHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /help
	tg.log.LogDebug("helplHander(): %s", update.Message.Text)
	msg := "/help - show all commands in admin menu (this command)\n" +
		"/info - show app info: app_name, app_IP, loglevel, autoforward, admins, etc.\n" +
		"/loglevel [LVL] - set level of logs to dev purposes (LVL: TRACE, DEBUG, INFO, WARN, ERROR, PANIC, FATAL, NOLOG(default))\n" +
		"/teststream - post test message (template) to admin and all users\n" +
		"/autoforward [FWD] - set forwarding mode for twich notification (FWD: DEBUG-admin channel only, OFF-admin and user channel, ON-send to all)\n" +
		"/autodel [DEL] - set autodelete for notification message when stream is offline (DEL: OFF, ON)\n" +
		"/post [MSG] - send any MSG as notification to admin and all users\n" +
		"/getadmins [ID] - show all admins in ID channel\n" +
		"/sendmsg [ID] [MSG] - post MSG in ID channel\n" +
		"/delall - delete all posted mesages\n" +
		"Forward from any autorized user to bot - forward any post or message into bot private chat if you have rights bot resends it to admin and all users\n"
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg,
	})
}

func (tg *Tg) TTVNotifyUserOnline(ttvUserID string, ttvStreams [][4]string) {
	tg.mu.Lock()
	defer tg.mu.Unlock()

	tg.log.LogDebug("TGnotify() Online:%s %v", ttvUserID, ttvStreams)
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
	msg := tgUser.Longname + " is online !\n\n" +
		ttvUser[1] + "  |  " + ttvUser[2] + "\n" +
		ttvUser[3] + "\n\n" +
		"https://www.twitch.tv/" + ttvUser[1]
	fileData, _ := tg.fs.ReadFile("data/" + ttvUser[1] + "_pic.jpg")
	_, errDEBUG := tg.bot.SendPhoto(tg.ctx, &TG.SendPhotoParams{
		ChatID:  tg.cfg.GetJsonAdmin().TgChannelID,
		Photo:   &TGm.InputFileUpload{Filename: ttvUser[1] + "_pic.jpg", Data: bytes.NewReader(fileData)},
		Caption: msg,
	})
	if errDEBUG != nil {
		tg.log.LogDebug("TTVnotify() DEBUG error: ChanID[%s]: %s", tg.cfg.GetJsonAdmin().TgChannelID, errDEBUG.Error())
		_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
			ChatID: tg.cfg.GetJsonAdmin().TgUserID,
			Text:   fmt.Sprintf("TTVUserOnlineNotify DEBUG() error: %s", errDEBUG.Error()),
		})
	}
	var (
		sentMsg *TGm.Message
		errON   error
	)
	switch tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) {
	case T.AFORW_ON:
		var idx int
		for i, el := range tg.cfg.GetJsonUsers() {
			if el.TgUserID == tgUser.TgUserID {
				sentMsg, errON = tg.bot.SendPhoto(tg.ctx, &TG.SendPhotoParams{
					ChatID:  tgUser.TgChannelID,
					Photo:   &TGm.InputFileUpload{Filename: ttvUser[1] + "_pic.jpg", Data: bytes.NewReader(fileData)},
					Caption: msg,
				})
				if errON != nil {
					tg.log.LogDebug("TTVUserOnlineNotify() FWD_ON error: ChanID[%s]: %s", tgUser.TgChannelID, errON.Error())
					_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
						ChatID: tg.cfg.GetJsonAdmin().TgUserID,
						Text:   fmt.Sprintf("TTVUserOnlineNotify() FWD_ON error: ChanID[%s]: %s", tgUser.TgChannelID, errON.Error()),
					})
				}
				if tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) == T.ADEL_ON {
					idx = i
					tg.msgsToDel[idx] = append(tg.msgsToDel[idx], delMsg{ChanID: tgUser.TgChannelID, MsgID: sentMsg.ID})
				}
			}
		}
		for _, el := range tg.cfg.GetJsonUsers() {
			if el.TgUserID != tgUser.TgUserID {
				go func() {
					fwdMsg, errON := tg.bot.ForwardMessage(tg.ctx, &TG.ForwardMessageParams{
						ChatID:     el.TgChannelID,
						FromChatID: tgUser.TgChannelID,
						MessageID:  sentMsg.ID,
					})
					if errON != nil {
						tg.log.LogDebug("TTVUserOnlineNotify() FWD_ON error: %s: ChanID[%s]", errON.Error(), el.TgChannelID)
						_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
							ChatID: tg.cfg.GetJsonAdmin().TgUserID,
							Text:   fmt.Sprintf("TTVUserOnlineNotify() FWD_ON error: ChanID[%s]: %s", tgUser.TgChannelID, errON.Error()),
						})
					}
					if tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) == T.ADEL_ON {
						tg.msgsToDel[idx] = append(tg.msgsToDel[idx], delMsg{ChanID: el.TgChannelID, MsgID: fwdMsg.ID})
					}
				}()
			}
		}
	case T.AFORW_OFF:
		sentMsg, errOFF := tg.bot.SendPhoto(tg.ctx, &TG.SendPhotoParams{
			ChatID:  tgUser.TgChannelID,
			Photo:   &TGm.InputFileUpload{Filename: ttvUser[1] + "_pic.jpg", Data: bytes.NewReader(fileData)},
			Caption: msg,
		})
		if errOFF != nil {
			tg.log.LogDebug("TTVnotify() FWD_OFF error: ChanID[%s]: %s", tgUser.TgChannelID, errOFF.Error())
			_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
				ChatID: tg.cfg.GetJsonAdmin().TgUserID,
				Text:   fmt.Sprintf("TTVUserOnlineNotify() FWD_OFF error: %s", errOFF.Error()),
			})
		}
		if tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) == T.ADEL_ON {
			for i, el := range tg.cfg.GetJsonUsers() {
				if el.TtvUserID == ttvUserID {
					tg.msgsToDel[i] = append(tg.msgsToDel[i], delMsg{ChanID: tgUser.TgChannelID, MsgID: sentMsg.ID})
					break
				}
			}
		}
	}
}

func (tg *Tg) TTVNotifyUserOffline(userID string, userName string, dur time.Duration) {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
		DisableNotification: true,
		ChatID:              tg.cfg.GetJsonAdmin().TgChannelID,
		Text:                fmt.Sprintf("%s went offline ~1h ago \nstream lasted ~%v \n", userName, dur),
	})
	tg.log.LogDebug("TGnotify() Offline: %s[%s]", userName, userID)
	if tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) == T.ADEL_ON {
		for i, el := range tg.cfg.GetJsonUsers() {
			if el.TtvUserID == userID {
				for _, elem := range tg.msgsToDel[i] {
					_, _ = tg.bot.DeleteMessage(tg.ctx, &TG.DeleteMessageParams{
						ChatID:    elem.ChanID,
						MessageID: elem.MsgID,
					})
				}
				break
			}
		}
	}
}

func (tg *Tg) delallHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /delall
	tg.log.LogDebug("delHander() %s", update.Message.Text)
	for i := range tg.cfg.GetJsonUsers() {
		for _, elem := range tg.msgsToDel[i] {
			_, _ = tg.bot.DeleteMessage(tg.ctx, &TG.DeleteMessageParams{
				ChatID:    elem.ChanID,
				MessageID: elem.MsgID,
			})
		}
	}
}

func (tg *Tg) postHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /post
	tg.log.LogDebug("postHander() %s", update.Message.Text)
	command := strings.Split(update.Message.Text, " ")
	if len(command) > 1 {
		var (
			sentMsg *TGm.Message
			err     error
		)
		msg := strings.Join(command[1:], " ")
		sentMsg, err = bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: tg.cfg.GetJsonAdmin().TgChannelID,
			Text:   msg,
		})
		if err != nil {
			tg.log.LogDebug("postHandler() DEBUG error: %s: ChanID[%s]", err.Error(), tg.cfg.GetJsonAdmin().TgChannelID)
			_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("command /post error: %s", err.Error()),
			})
		}
		userID := fmt.Sprintf("%d", update.Message.From.ID)
		var channelID string
		for _, el := range tg.cfg.GetJsonUsers() {
			if el.TgUserID == userID {
				channelID = el.TgChannelID
				sentMsg, err = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
					ChatID: channelID,
					Text:   msg,
				})
				if err != nil {
					tg.log.LogDebug("command /post error: ChanID[%s]: %s", el.TgChannelID, err.Error())
					_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Tcommand /post error: ChanID[%s]: %s", el.TgChannelID, err.Error()),
					})
				}
			}
		}
		for _, el := range tg.cfg.GetJsonUsers() {
			if el.TgUserID != userID {
				go func() {
					_, err = tg.bot.ForwardMessage(tg.ctx, &TG.ForwardMessageParams{
						ChatID:     el.TgChannelID,
						FromChatID: channelID,
						MessageID:  sentMsg.ID,
					})
					if err != nil {
						tg.log.LogDebug("command /post error: %s: ChanID[%s]", err.Error(), el.TgChannelID)
						_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
							ChatID: tg.cfg.GetJsonAdmin().TgUserID,
							Text:   fmt.Sprintf("command /post error: ChanID[%s]: %s", el.TgChannelID, err.Error()),
						})
					}
				}()
			}
		}
		return
	}
	tg.log.LogDebug("postHander(): should use Message as 1 flag")
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command /post should use Message as 1 flag",
	})
}

func (tg *Tg) defaultHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // repost
	tg.log.LogDebug("defaultHander(): ID:%d TEXT:%s", update.Message.ID, update.Message.Text)
	if update.Message.ForwardOrigin != nil {
		_, err := bot.ForwardMessage(ctx, &TG.ForwardMessageParams{
			ChatID:     tg.cfg.GetJsonAdmin().TgChannelID,
			FromChatID: update.Message.From.ID,
			MessageID:  update.Message.ID,
		})
		if err != nil {
			tg.log.LogDebug("defaultHandler() error: ChanID[%s]: %s", tg.cfg.GetJsonAdmin().TgChannelID, err.Error())
			_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("defaultHandler() forward error: %s", err.Error()),
			})
		}
		channelID := fmt.Sprintf("%d", update.Message.From.ID)
		for _, el := range tg.cfg.GetJsonUsers() {
			if el.TgChannelID != channelID {
				go func() {
					_, err := bot.ForwardMessage(ctx, &TG.ForwardMessageParams{
						ChatID:     el.TgChannelID,
						FromChatID: update.Message.From.ID,
						MessageID:  update.Message.ID,
					})
					if err != nil {
						tg.log.LogDebug("defaultHandler() forward error: ChanID[%s] error: %s", el.TgChannelID, err.Error())
						_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
							ChatID: tg.cfg.GetJsonAdmin().TgUserID,
							Text:   fmt.Sprintf("defaultHandler() forward error: %s ID:%s", err.Error(), el.TgChannelID),
						})
					}
				}()
			}
		}
		return
	}
	tg.log.LogDebug("defaultHander(): command not found")
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "command not found",
	})
}
