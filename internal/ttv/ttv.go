package ttv

import (
	"context"
	"fmt"
	T "team_streams/internal/types"
	"time"

	TTV "github.com/nicklaw5/helix/v2"
)

var _ T.ITTV = (*Ttv)(nil)

const (
	TICK_DURATION   = 1 * time.Minute
	OFFLINE_COUNTER = 60 // (* TICK_DURATION)
)

type Ttv struct {
	log          T.ILog
	cfg          T.ICfg
	tg           T.ITG
	ticker       *time.Ticker
	userIDs      []string
	userNames    []string
	offlineUsers []int
	onlineUsers  []time.Time
}

func NewTTVApp(cfg T.ICfg, log T.ILog, tg T.ITG) *Ttv {
	alen := len(cfg.GetJsonUsers())
	offlineUsers := make([]int, alen)
	onlineUsers := make([]time.Time, alen)
	userIDs := make([]string, 0, alen)
	userNames := make([]string, 0, alen)
	for _, user := range cfg.GetJsonUsers() {
		userIDs = append(userIDs, user.TtvUserID)
		userNames = append(userNames, user.Nickname)
	}
	return &Ttv{
		log:          log,
		cfg:          cfg,
		tg:           tg,
		ticker:       time.NewTicker(TICK_DURATION),
		userIDs:      userIDs,
		userNames:    userNames,
		offlineUsers: offlineUsers,
		onlineUsers:  onlineUsers,
	}
}

func (ttv *Ttv) clientGetStreams(ttvClient *TTV.Client) {
	var (
		err        error
		respToken  *TTV.AppAccessTokenResponse
		respStream *TTV.StreamsResponse
	)

LabelStart:
	respStream, err = ttvClient.GetStreams(&TTV.StreamsParams{
		UserIDs: ttv.userIDs,
	})
	if err != nil {
		ttv.log.LogError(fmt.Errorf("%s: %w", "ttvClient.GetStreams() could not get StreamsData: ", err))
		return
	}
	if respStream.StatusCode == 401 {
		respToken, err = ttvClient.RequestAppAccessToken([]string{})
		if err != nil {
			ttv.log.LogError(fmt.Errorf("%s: %w", "ttvClient.RequestAppAccessToken() could not get AccessToken: ", err))
			return
		}
		ttv.cfg.SetEnvVal(T.TTV_APPACCESS_TOKEN, respToken.Data.AccessToken)
		ttvClient.SetAppAccessToken(respToken.Data.AccessToken)
		goto LabelStart
	}
	ttvStreams := make([][4]string, 0, len(ttv.cfg.GetJsonUsers()))
	for _, elem := range respStream.Data.Streams {
		ttvStreams = append(ttvStreams, [4]string{elem.UserID, elem.UserLogin, elem.GameName, elem.Title})
	}
LabelUserOnline:
	for i, el := range ttv.userIDs {
		for _, elem := range respStream.Data.Streams {
			if el == elem.UserID {
				if ttv.onlineUsers[i].IsZero() {
					ttv.onlineUsers[i] = time.Now()
					ttv.offlineUsers[i] = 0
					ttv.log.LogInfo("ttvClient.GetStreams(): %s[%s] online, %v", elem.UserLogin, elem.UserID, ttvStreams)
					go ttv.tg.TTVNotifyUserOnline(elem.UserID, ttvStreams)
				}
				continue LabelUserOnline
			}
		}
		if !ttv.onlineUsers[i].IsZero() {
			if ttv.offlineUsers[i] < OFFLINE_COUNTER {
				ttv.offlineUsers[i]++
			} else {
				ttv.log.LogInfo("ttvClient.GetStreams(): %s[%s] offline", ttv.userNames[i], ttv.userIDs[i])
				go ttv.tg.TTVNotifyUserOffline(ttv.userIDs[i], ttv.userNames[i], time.Since(ttv.onlineUsers[i].Add(time.Hour)).Truncate(time.Minute))
				ttv.onlineUsers[i] = time.Time{}
				ttv.offlineUsers[i] = 0
			}
		}
	}
}

func (ttv *Ttv) clientUpdate(ctx context.Context) {
	ttv.ticker.Reset(TICK_DURATION)
	ttvClient, err := TTV.NewClient(&TTV.Options{
		ClientID:     ttv.cfg.GetEnvVal(T.TTV_CLIENT_ID),
		ClientSecret: ttv.cfg.GetEnvVal(T.TTV_CLIENT_SECRET),
	})
	if err != nil {
		ttv.log.LogError(fmt.Errorf("%s: %w", "ttvClient could not be created: ", err))
	}
	ttvClient.SetAppAccessToken(ttv.cfg.GetEnvVal(T.TTV_APPACCESS_TOKEN))

	go func() {
	LabelUpdateStop:
		for {
			select {
			case <-ctx.Done():
				break LabelUpdateStop
			case <-ttv.ticker.C:
				ttv.clientGetStreams(ttvClient)
			}
		}
		ttv.ticker.Stop()
	}()
}

func (ttv *Ttv) Start() func(err error) {
	ctxTTVUpdate, ctxCancelTTVUpdate := context.WithCancel(context.Background())
	ttv.clientUpdate(ctxTTVUpdate)
	ttv.log.LogInfo("TTV_app started")
	return func(err error) { // TtvStop
		ctxCancelTTVUpdate()
		if err != nil {
			ttv.log.LogError(fmt.Errorf("%s: %w", "TTV_app stoped with error", err))
		} else {
			ttv.log.LogInfo("TTV_app stoped")
		}
	}
}
