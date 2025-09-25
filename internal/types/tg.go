package types

type ITG interface {
	Start() func(err error)
}
