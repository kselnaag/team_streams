package app

import (
	"fmt"
	"os"
	"path/filepath"

	C "team_streams/internal/cfg"
	L "team_streams/internal/log"
	T "team_streams/internal/types"

	TG "team_streams/internal/tg"
	TTV "team_streams/internal/ttv"
)

type App struct {
	appname string
	cfg     T.ICfg
	log     T.ILog
	ttv     T.ITTV
	tg      T.ITG
}

func execPathAndFname() (string, string) {
	path, _ := os.Executable()
	return filepath.Split(path)
}

func NewApp() *App {
	appdir, appname := execPathAndFname()
	cfg := C.NewCfgMap(appdir, appname).Parse()
	log := L.NewLogFprintf(cfg, 0, 0)
	ttv := TTV.NewTTVApp(cfg, log)
	tg := TG.NewTGBot(cfg, log, ttv)
	return &App{
		appname: appname,
		cfg:     cfg,
		log:     log,
		ttv:     ttv,
		tg:      tg,
	}
}

func (a *App) Start() func(err error) {
	logStop := a.log.Start()
	a.log.LogInfo(a.appname + " app started")

	// os.Exit(0)
	return func(err error) { // AppStop

		if err != nil {
			a.log.LogPanic(fmt.Errorf("%s: %w", a.appname+" app stoped with error", err))
		} else {
			a.log.LogInfo(a.appname + " app stoped")
		}
		logStop()
	}
}
