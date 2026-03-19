package types

type ITTV interface {
	Start() func(err error)
}

type StreamInfoTTV struct {
	UserID    string
	UserLogin string
	Game      string
	Title     string
}
