package types

type ITTV interface {
	Start() func(err error)
}
