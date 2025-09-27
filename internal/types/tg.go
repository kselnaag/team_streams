package types

type ITG interface {
	Start() func(error)
	TTVUserOnlineNotify(string, [][4]string)
}
