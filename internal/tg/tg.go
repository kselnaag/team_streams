package tg

import (
	T "team_streams/internal/types"

	_ "github.com/go-telegram/bot"
	_ "github.com/go-telegram/bot/models"
)

var _ T.ITG = (*Tg)(nil)

type Tg struct {
}

func NewTGBot(cfg T.ICfg, log T.ILog, ttv T.ITTV) *Tg {

	return &Tg{}
}
