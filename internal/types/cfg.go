package types

type ICfg interface {
	GetEnvVal(string) string
	SetEnvVal(string, string)
	GetJsonVals() []User
	Parse() ICfg
}

const (
	TS_APP_NAME         = "TS_APP_NAME"
	TS_APP_IP           = "TS_APP_IP"
	TS_LOG_LEVEL        = "TS_LOG_LEVEL"
	TS_APP_AUTOFORWARD  = "TS_APP_AUTOFORWARD"
	TG_BOT_TOKEN        = "TG_BOT_TOKEN"
	TTV_CLIENT_ID       = "TTV_CLIENT_ID"
	TTV_CLIENT_SECRET   = "TTV_CLIENT_SECRET"
	TTV_APPACCESS_TOKEN = "TTV_APPACCESS_TOKEN"
)

type User struct {
	Nickname    string
	Shortname   string
	Longname    string
	TtvUserID   string
	TgUserID    string
	TgChannelID string
}

type Users struct {
	Users []User
}
