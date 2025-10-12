package types

type ICfg interface {
	GetEnvVal(string) string
	SetEnvVal(string, string)
	GetJsonUsers() []User
	GetJsonAdmin() User
	Parse() ICfg
}

const (
	TS_APP_NAME         = "TS_APP_NAME"
	TS_APP_IP           = "TS_APP_IP"
	TS_LOG_LEVEL        = "TS_LOG_LEVEL"
	TS_APP_AUTOFORWARD  = "TS_APP_AUTOFORWARD"
	TS_APP_AUTODEL      = "TS_APP_AUTODEL"
	TG_BOT_TOKEN        = "TG_BOT_TOKEN"
	TTV_CLIENT_ID       = "TTV_CLIENT_ID"
	TTV_CLIENT_SECRET   = "TTV_CLIENT_SECRET"
	TTV_APPACCESS_TOKEN = "TTV_APPACCESS_TOKEN"
)

type User struct {
	Nickname    string
	Longname    string
	TtvUserID   string
	TgUserID    string
	TgChannelID string
	TgChatID    string
}

type JsonVals struct {
	Admin User
	Users []User
}
