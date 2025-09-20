package ttv

import (
	T "team_streams/internal/types"

	_ "github.com/nicklaw5/helix/v2"
)

var _ T.ITTV = (*Ttv)(nil)

type Ttv struct {
}

func NewTTVApp(cfg T.ICfg, log T.ILog) *Ttv {

	return &Ttv{}
}
