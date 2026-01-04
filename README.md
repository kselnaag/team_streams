# **TEAM_STREAMS**<br>📱 Integration bot for coupling twitch and telegram channels 💻 
----

## 🍱 System parts
- team_streams_app - TwitchAPI_app as TTV integration part *(make it in Twitch dev panel)*
- team_streams_bot - TelegramAPI_bot as TG integration part *(make it throught BotFather in Telegram)*
- team_streams - external service with logic and usecases *(this project)*

## ⚡ Features
- Just start TTV stream: bot makes post about it in main TG channel and forwards it to another team members
- Make post manualy in your TG channel and forward it into bot: it forwards post to another team members
- Make post in bot privat chat: bot makes this post in your TG channel and forwards it to another team members
- Bot options control throught bot private chat by special commands

You should add the bot to all TG chanels as administrator with posting rights and start it.

## 📜 Configs

Can pass tokens throught env vars: TG_BOT_TOKEN, TTV_CLIENT_ID, TTV_CLIENT_SECRET required

🔒 App credentials: 🔑
```
kselnaag:~/team_streams$ cat ./bin/team_streams.env
# LOG levels: TRACE, DEBUG, INFO, WARN, ERROR, PANIC, FATAL, NOLOG(default if empty or mess)
TS_LOG_LEVEL=INFO
TS_APP_IP=localhost
# TTV AUTOFORWARD: DEBUG, ON, OFF
TS_APP_AUTOFORWARD=DEBUG
# TG AUTODEL: OFF, ON
TS_APP_AUTODEL=OFF
TG_BOT_TOKEN=
TTV_CLIENT_ID=
TTV_CLIENT_SECRET=
TTV_APPACCESS_TOKEN=
```

👥 Team members: 👥
```
kselnaag:~/team_streams$ cat ./bin/team_streams.json
{"admin":{
    "nickname":"",
    "longname":"",
    "ttvUserID":"",
    "tgUserID":"",
    "tgChannelID":"",
    "tgChatID":""
    },
"users":[{
    "nickname":"",
    "longname":"",
    "ttvUserID":"",
    "tgUserID":"",
    "tgChannelID":"",
    "tgChatID":""
    },{
    "nickname":"",
    "longname":"",
    "ttvUserID":"",
    "tgUserID":"",
    "tgChannelID":"",
    "tgChatID":""
    },{
    "nickname":"",
    "longname":"",
    "ttvUserID":"",
    "tgUserID":"",
    "tgChannelID":"",
    "tgChatID":""
    }
]}
```
Fill and save this configs near by the executable file

📂 Start folder: 🏁
```
kselnaag:~/team_streams/bin$ ll
drwxrwxrwx 1 ksel ksel    4096 Sep 23 04:38 ./
drwxrwxrwx 1 ksel ksel    4096 Sep 22 23:28 ../
-rwxrwxrwx 1 ksel ksel 9526414 Oct  1 22:04 stoilo_streams*
-rwxrwxrwx 1 ksel ksel     347 Oct  1 21:50 team_streams.env
-rwxrwxrwx 1 ksel ksel      28 Oct  1 21:50 team_streams.json
```
To re-read configs without stopping process use `kill -SIGHUP <pid>` (server access required) or use options control throught bot private chat (if authorized in TG)

## ⚙️ Build script

```
kselnaag:~/team_streams$ go version
go version go1.25.1 linux/amd64

kselnaag:~/team_streams$ ./build/build.sh
+ GOOS=linux
+ GOARCH=amd64
+ go build -o ./bin/team_streams ./cmd/main.go

kselnaag:~/team_streams$ ldd ./bin/team_streams
        linux-vdso.so.1 (0x00007ffdc2406000)
        libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x000073cc27400000)
        /lib64/ld-linux-x86-64.so.2 (0x000073cc276ca000)
```
In case of emergency change the build script

## 💡 Bot operations description

Commands in bot private chat, for registered users only:
- /help - show all commands in admin menu
- /info - show app info: app_name, app_IP, loglevel, autoforward, admins, etc.
-	/loglevel [LVL] - set level of logs to dev purposes (LVL: TRACE, DEBUG, INFO, WARN, ERROR, PANIC, FATAL, NOLOG(default))
-	/teststream - post test message (template) to admin and all users
-	/autoforward [FWD] - set forwarding mode for twich notification (FWD: DEBUG-admin channel only, OFF-admin and user channel, ON-send to all)
-	/autodel [DEL] - set autodelete for notification message when stream is offline (DEL: OFF, ON)
-	/post [MSG] - send any MSG as notification to admin and all users
-	/getadmins [ID] - show all admins in ID channel
-	/sendmsg [ID] [MSG] - post MSG in ID channel
-	/delall - delete all posted mesages
- "Forward from any autorized user to bot" - forward any post or message into bot private chat if you have rights bot resends it to admin and all users

## 🦋 Inspired by STOILO_TEAM

<p align="center">
<br>
|
  <a href="https://www.twitch.tv/dayopponent" title="https://www.twitch.tv/dayopponent" >dayopponent</a> |
  <a href="https://www.twitch.tv/iksssy" title="https://www.twitch.tv/iksssy">iksssy</a> |
  <a href="https://www.twitch.tv/mewendi" title="https://www.twitch.tv/mewendi">mewendi</a>
|
<br><br>
<img style="margin-right: 50px;" width="20%" src="pic/dayopponent.jpg" title="dayopponent" alt="dayopponent">
<img style="margin-bottom: 5px;" width="20%" src="pic/iksssy.jpg" title="iksssy" alt="iksssy">
<img style="margin-left: 50px;"width="20%" src="pic/mewendi.jpg" title="mewendi" alt="mewendi">
</p>
<br><br>

----
### **🔗 LINKS**
| [TG_bot lib](github.com/go-telegram/bot "github.com/go-telegram/bot")
| [TTV_app lib](github.com/nicklaw5/helix "github.com/nicklaw5/helix")
|
