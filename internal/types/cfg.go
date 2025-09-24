package types

type ICfg interface {
	GetEnvVal(string) string
	SetEnvVal(string, string)
	// GetJsonVal(string) string
	Parse() ICfg
}

const (
	TS_APP_NAME         = "SS_APP_NAME"
	TS_APP_IP           = "SS_APP_IP"
	TS_LOG_LEVEL        = "SS_LOG_LEVEL"
	TG_BOT_TOKEN        = "TG_BOT_TOKEN"
	TTV_CLIENT_ID       = "TTV_CLIENT_ID"
	TTV_CLIENT_SECRET   = "TTV_CLIENT_SECRET"
	TTV_APPACCESS_TOKEN = "TTV_APPACCESS_TOKEN"
)
