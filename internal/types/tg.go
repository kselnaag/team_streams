package types

type ITG interface {
	Start() func(error)
	TTVUserOnlineNotify(string, [][4]string)
}

const (
	COMMAND_INFO        = "info"
	COMMAND_LOGLEVEL    = "loglevel"
	COMMAND_TESTSTREAM  = "teststream"
	COMMAND_AUTOFORWARD = "autoforward"
	COMMAND_POST        = "post"
	COMMAND_GETADMINS   = "getadmins"
	COMMAND_SENDMSG     = "sendmsg"
)
