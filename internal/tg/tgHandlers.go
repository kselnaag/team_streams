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

func (tg *Tg) TTVNotifyUserOnline(ttvStream T.StreamInfoTTV) {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	tg.log.LogDebug("TGnotify() Online: %v", ttvStream)
	var tgUser T.User
	for _, el := range tg.cfg.GetJsonUsers() {
		if el.TtvUserID == ttvStream.UserID {
			tgUser = el
			break
		}
	}
	chatUrl := "https://t.me/"
	if chat, errCh := tg.bot.GetChat(tg.ctx, &TG.GetChatParams{ChatID: tgUser.TgChatID}); errCh == nil {
		chatUrl += chat.Username
	}
	msg := tgUser.Longname + " уже запустил(а) стрим!" + "\n" + "https://twitch.tv/" + ttvStream.UserLogin
	notifyKeyboard := [][]TGm.InlineKeyboardButton{{{Text: "В ТГ ГРУППУ", URL: chatUrl}, {Text: "НА СТРИМ", URL: "https://twitch.tv/" + ttvStream.UserLogin}}}
	fileData, _ := tg.fs.ReadFile("data/" + ttvStream.UserLogin + "_pic.jpg")
	fileUpload := TGm.InputFileUpload{Filename: ttvStream.UserLogin + "_pic.jpg", Data: bytes.NewReader(fileData)}
	photoParams := TG.SendPhotoParams{
		ChatID:                tg.cfg.GetJsonAdmin().TgChannelID,
		Photo:                 &fileUpload,
		Caption:               msg,
		ShowCaptionAboveMedia: true,
		ReplyMarkup:           TGm.InlineKeyboardMarkup{InlineKeyboard: notifyKeyboard},
	}
	if _, errDEBUG := tg.bot.SendPhoto(tg.ctx, &photoParams); errDEBUG != nil { // TS_APP_AUTOFORWARD == "DEBUG"
		tg.log.LogDebug("TTVnotify() DEBUG error: ChanID[%s]: %s", tg.cfg.GetJsonAdmin().TgChannelID, errDEBUG.Error())
		_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
			ChatID: tg.cfg.GetJsonAdmin().TgUserID,
			Text:   fmt.Sprintf("TTVUserOnlineNotify DEBUG() error: %s", errDEBUG.Error()),
		})
	}
	var (
		sentMsg, fwdMsg *TGm.Message
		errOFF, errON   error
	)
	if (tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) == T.ADEL_OFF) || (tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) == T.AFORW_ON) { // TS_APP_AUTOFORWARD == "OFF"
		fileData, _ = tg.fs.ReadFile("data/" + ttvStream.UserLogin + "_pic.jpg")
		fileUpload = TGm.InputFileUpload{Filename: ttvStream.UserLogin + "_pic.jpg", Data: bytes.NewReader(fileData)}
		photoParams = TG.SendPhotoParams{
			ChatID:                tgUser.TgChannelID,
			Photo:                 &fileUpload,
			Caption:               msg,
			ShowCaptionAboveMedia: true,
			ReplyMarkup:           TGm.InlineKeyboardMarkup{InlineKeyboard: notifyKeyboard},
		}
		if sentMsg, errOFF = tg.bot.SendPhoto(tg.ctx, &photoParams); errOFF != nil {
			tg.log.LogDebug("TTVUserOnlineNotify() FWD_OFF error: %s[%s]: %s", tgUser.Nickname, tgUser.TgChannelID, errOFF.Error())
			_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
				ChatID: tg.cfg.GetJsonAdmin().TgUserID,
				Text:   fmt.Sprintf("TTVUserOnlineNotify() FWD_OFF error: %s[%s]: %s", tgUser.Nickname, tgUser.TgChannelID, errOFF.Error()),
			})
		} else {
			if tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) == T.ADEL_ON {
				for idx, el := range tg.cfg.GetJsonUsers() {
					if el.TtvUserID == ttvStream.UserID {
						tg.msgsToDel[idx] = append(tg.msgsToDel[idx], delMsg{ChanID: tgUser.TgChannelID, MsgID: sentMsg.ID})
						break
					}
				}
			}
		}
	}
	if tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) == T.AFORW_ON { // TS_APP_AUTOFORWARD == "ON"
		uniqueChannels := make(map[string]struct{}, 8)
		for idx, el := range tg.cfg.GetJsonUsers() {
			if _, ok := uniqueChannels[el.TgChannelID]; !ok && (el.TgUserID != tgUser.TgUserID) {
				uniqueChannels[el.TgChannelID] = struct{}{}
				fwdMsg, errON = tg.bot.ForwardMessage(tg.ctx, &TG.ForwardMessageParams{
					ChatID:     el.TgChannelID,
					FromChatID: tgUser.TgChannelID,
					MessageID:  sentMsg.ID,
				})
				if errON != nil {
					tg.log.LogDebug("TTVUserOnlineNotify() FWD_ON error: %s[%s]: %s", el.Nickname, el.TgChannelID, errON.Error())
					_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
						ChatID: tg.cfg.GetJsonAdmin().TgUserID,
						Text:   fmt.Sprintf("TTVUserOnlineNotify() FWD_ON error: %s[%s]: %s", el.Nickname, el.TgChannelID, errON.Error()),
					})
				} else {
					if tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) == T.ADEL_ON {
						tg.msgsToDel[idx] = append(tg.msgsToDel[idx], delMsg{ChanID: el.TgChannelID, MsgID: fwdMsg.ID})
					}
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
		Text:                fmt.Sprintf("%s went offline ~1h ago \nstream lasted ~%v", userName, dur),
	})
	tg.log.LogDebug("TGnotify() Offline: %s[%s]", userName, userID)
	if tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) == T.ADEL_ON {
		for idx, el := range tg.cfg.GetJsonUsers() {
			if el.TtvUserID == userID {
				for _, elem := range tg.msgsToDel[idx] {
					_, _ = tg.bot.DeleteMessage(tg.ctx, &TG.DeleteMessageParams{
						ChatID:    elem.ChanID,
						MessageID: elem.MsgID,
					})
				}
				tg.msgsToDel[idx] = tg.msgsToDel[idx][:0]
				break
			}
		}
	}
}

func (tg *Tg) infoHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /info
	tg.log.LogDebug("infolHander(): %s", update.Message.Text)
	cfgmsg := T.TS_APP_NAME + "=" + tg.cfg.GetEnvVal(T.TS_APP_NAME) + "\n" +
		T.TS_APP_IP + "=" + tg.cfg.GetEnvVal(T.TS_APP_IP) + "\n" +
		T.TS_APP_AUTOFORWARD + "=" + tg.cfg.GetEnvVal(T.TS_APP_AUTOFORWARD) + "\n" +
		T.TS_APP_AUTODEL + "=" + tg.cfg.GetEnvVal(T.TS_APP_AUTODEL) + "\n" +
		T.TS_LOG_LEVEL + "=" + tg.cfg.GetEnvVal(T.TS_LOG_LEVEL) + "\n"
	admins := "Admins:\n" + fmt.Sprintf("%v", tg.getChatAdmins(tg.cfg.GetJsonAdmin().TgChannelID, update)) + "\n"
	channels := "Channels:\n" + fmt.Sprintf("%v", tg.getChanInfo(ctx, bot)) + "\n"
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   cfgmsg + admins + channels,
	})
}

func (tg *Tg) helpHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /help
	tg.log.LogDebug("helplHander(): %s", update.Message.Text)
	msg := "/help - show all commands in admin menu (this command)\n" +
		"/info - show app info: app_name, app_IP, loglevel, autoforward, admins, etc.\n" +
		"/loglevel [LVL] - set level of logs to dev purposes (LVL: TRACE, DEBUG, INFO, WARN, ERROR, PANIC, FATAL, NOLOG(default))\n" +
		"/autoforward [FWD] - set forwarding mode for twich notification (FWD: DEBUG-admin channel only, OFF-admin and user channel, ON-send to all)\n" +
		"/autodel [DEL] - set autodelete for notification message when stream is offline (DEL: OFF, ON)\n" +
		"/delall - delete all posted mesages\n"
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg,
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

func (tg *Tg) delallHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) { // /delall
	tg.log.LogDebug("delHander() %s", update.Message.Text)
	for i := range tg.cfg.GetJsonUsers() {
		for _, elem := range tg.msgsToDel[i] {
			_, _ = tg.bot.DeleteMessage(tg.ctx, &TG.DeleteMessageParams{
				ChatID:    elem.ChanID,
				MessageID: elem.MsgID,
			})
		}
		tg.msgsToDel[i] = tg.msgsToDel[i][:0]
	}
}
