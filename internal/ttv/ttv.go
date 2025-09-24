package ttv

import (
	"fmt"
	T "team_streams/internal/types"
	"time"

	TTV "github.com/nicklaw5/helix/v2"
)

var _ T.ITTV = (*Ttv)(nil)

type Ttv struct {
	log       T.ILog
	cfg       T.ICfg
	tg        T.ITG
	tickShort *time.Ticker
	tickLong  *time.Ticker
}

func NewTTVApp(cfg T.ICfg, log T.ILog, tg T.ITG) *Ttv {
	return &Ttv{
		log:       log,
		cfg:       cfg,
		tg:        tg,
		tickShort: time.NewTicker(10 * time.Second),
		tickLong:  time.NewTicker(1 * time.Minute),
	}
}

func (ttv *Ttv) ClientReconnect() *TTV.Client {
	var (
		ttvClient *TTV.Client
		err       error
	)
	for range ttv.tickShort.C {
		ttvClient, err = TTV.NewClient(&TTV.Options{
			ClientID:     ttv.cfg.GetEnvVal(T.TTV_CLIENT_ID),
			ClientSecret: ttv.cfg.GetEnvVal(T.TTV_CLIENT_SECRET),
		})
		if err != nil {
			ttv.log.LogWarn("%s", "ttvClient could not connect: "+err.Error())
		} else {
			break
		}
	}
	return ttvClient
}

func (ttv *Ttv) Start() func(err error) {
	var (
		err        error
		okToken    bool
		respToken  *TTV.AppAccessTokenResponse
		respStream *TTV.StreamsResponse
	)
	ttvClient := ttv.ClientReconnect()
	for range ttv.tickLong.C {
		okToken, _, err = ttvClient.ValidateToken(ttv.cfg.GetEnvVal(T.TTV_APPACCESS_TOKEN)) // TTVclient.ValidateToken(accessToken string) (bool, *ValidateTokenResponse, error)
		if /*-ttvClient*/ err {
			ttvClient = ttv.ClientReconnect()
			continue
		}
		if !okToken {
			respToken, err = ttvClient.RequestAppAccessToken([]string{"user:read:email"})
			if err != nil {
				ttv.log.LogWarn("%s", "ttvClient.RequestAppAccessToken() could not get AccessToken: "+err.Error())
				continue
			}
			ttv.cfg.SetEnvVal(T.TTV_APPACCESS_TOKEN, respToken.Data.AccessToken)
		}

		respStream, err = ttvClient.GetStreams(&TTV.StreamsParams{
			UserLogins: []string{},
		})
		if err != nil {
			ttv.log.LogWarn("%s", "ttvClient.GetStreams() could not get StreamsData: "+err.Error())
			continue
		}
		for i := 0; i < len(respStream.Data.Streams); i++ {
			name := respStream.Data.Streams[i].UserLogin
			game := respStream.Data.Streams[i].GameName
			title := respStream.Data.Streams[i].Title
			fmt.Println(name + "\n" + game + "\n" + title + "\n")
		}

	}

	ttv.log.LogInfo("TTV_app started")
	return func(err error) { // TtvStop
		ttv.tickLong.Stop()
		ttv.tickShort.Stop()
		if err != nil {
			ttv.log.LogError(fmt.Errorf("%s: %w", "TTV_app stoped with error", err))
		} else {
			ttv.log.LogInfo("TTV_app stoped")
		}
	}
}
