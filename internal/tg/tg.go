package tg

import (
	"fmt"
	T "team_streams/internal/types"

	_ "github.com/go-telegram/bot"
	_ "github.com/go-telegram/bot/models"
)

var _ T.ITG = (*Tg)(nil)

type Tg struct {
	cfg T.ICfg
	log T.ILog
}

func NewTGBot(cfg T.ICfg, log T.ILog) *Tg {

	return &Tg{
		cfg: cfg,
		log: log,
	}
}

func (tg *Tg) Start() func(err error) {
	//tg.cfg.GetJsonVals

	tg.log.LogInfo("TG_app started")
	return func(err error) { // TgStop

		if err != nil {
			tg.log.LogError(fmt.Errorf("%s: %w", "TG_app stoped with error", err))
		} else {
			tg.log.LogInfo("TG_app stoped")
		}
	}
}
