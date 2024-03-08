package flysend

import (
	"context"
	"github.com/einherij/pilot/pkg/wsclient"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/einherij/pilot/pkg/flymap"
	"github.com/einherij/pilot/pkg/navigator"
)

type Sender struct {
	wsClient *wsclient.Client
	flyMap   *flymap.FlyMap
	nav      *navigator.Navigator
}

func New(wsClient *wsclient.Client, flyMap *flymap.FlyMap, nav *navigator.Navigator) *Sender {
	return &Sender{
		wsClient: wsClient,
		flyMap:   flyMap,
		nav:      nav,
	}
}

func (s *Sender) Run(ctx context.Context) {
	logrus.Warnf("starting fly map sender")
	flyMapTicker := time.NewTicker(time.Second)
	posTicker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-flyMapTicker.C:
			s.wsClient.SendMessage(wsclient.Message{
				Type:    wsclient.MTFlyMap,
				Content: s.flyMap.GetOBJ(),
			})
		case <-posTicker.C:
			s.wsClient.SendMessage(wsclient.Message{
				Type:    wsclient.MTPos,
				Content: s.nav.GetPos().GetOBJ(),
			})
		case <-ctx.Done():
			logrus.Warnf("stopped fly map sender")
			return
		}
	}
}
