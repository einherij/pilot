package main

import (
	"context"
	"github.com/SMerrony/tello"
	"github.com/sirupsen/logrus"
	"os"
	"time"

	"github.com/einherij/enterprise"
	"github.com/einherij/enterprise/utils"
	"github.com/einherij/pilot/pkg/controller"
	"github.com/einherij/pilot/pkg/flymap"
	"github.com/einherij/pilot/pkg/flymap/flysend"
	"github.com/einherij/pilot/pkg/navigator"
	"github.com/einherij/pilot/pkg/videosender"
	"github.com/einherij/pilot/pkg/wsclient"
)

func main() {
	handlerHostURL := os.Getenv("HANDLER_HOST_URL")

	app := enterprise.NewApplication()

	// connect to interface
	wsClient := wsclient.New(handlerHostURL)
	app.RegisterRunner(wsClient)

	d := new(tello.Tello)

	utils.PanicOnError(d.ControlConnectDefault())
	app.RegisterOnShutdown(func() {
		d.ControlDisconnect()
		logrus.Warnf("control disconnected")
	})

	// Video
	videoStream := utils.Must(d.VideoConnectDefault())
	app.RegisterOnShutdown(func() {
		d.VideoDisconnect()
		logrus.Warnf("video disconnected")
	})
	d.SetVideoWide()
	d.SetSportsMode(true)

	// send video key frames
	app.RegisterRunner(runnerFunc(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
				d.GetVideoSpsPps()
			}
		}
	}))

	videoSender := videosender.New(handlerHostURL, videoStream, videosender.StreamPipe, false)
	app.RegisterRunner(videoSender)

	// FlightData
	fdStream := utils.Must(d.StreamFlightData(false, 100))

	// position
	nav := navigator.NewNavigator(fdStream)
	app.RegisterRunner(nav)

	// map
	flyMap := flymap.New("FlyMap", "map.mtl")
	app.RegisterOnShutdown(func() { _ = flymap.SaveMap("./maps/map.obj", flyMap) })

	mapSender := flysend.New(wsClient, flyMap, nav)
	app.RegisterRunner(mapSender)

	cmdHandler := controller.New(wsClient, d, flyMap)
	app.RegisterRunner(cmdHandler)

	app.Run()
}

// TODO: add runnerFunc to enterprise
type runnerFunc func(ctx context.Context)

func (r runnerFunc) Run(ctx context.Context) {
	r(ctx)
}
