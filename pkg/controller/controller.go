package controller

import (
	"context"
	"fmt"
	"github.com/SMerrony/tello"
	"github.com/einherij/pilot/pkg/flymap"
	"github.com/einherij/pilot/pkg/vector"
	"github.com/einherij/pilot/pkg/wsclient"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	wsClient *wsclient.Client
	drone    *tello.Tello
	flyMap   *flymap.FlyMap
}

func New(wsClient *wsclient.Client, drone *tello.Tello, flyMap *flymap.FlyMap) *Controller {
	return &Controller{
		wsClient: wsClient,
		drone:    drone,
		flyMap:   flyMap,
	}
}

func (h *Controller) Run(ctx context.Context) {
	logrus.Warnf("started drone controller")
	var home vector.V3D
	var homeYaw int16
	var lastCheckpoint int
	for {
		select {
		case <-ctx.Done():
			logrus.Warnf("stopped drone controller")
			return
		default:
			msg := h.wsClient.ReceiveMessage(ctx)
			if msg.Type != wsclient.MTCmd {
				continue
			}
			fd := h.drone.GetFlightData()
			var info string
			switch string(msg.Content) {
			case "Dq":
				info = "Started Turning Left"
				h.drone.TurnLeft(100)
			case "Uq":
				info = "Stopped Turning Left"
				h.drone.Hover()
			case "De":
				info = "Started Turning Right"
				h.drone.TurnRight(100)
			case "Ue":
				info = "Stopped Turning Right"
				h.drone.Hover()
			case "Dw":
				info = "Started Going Forward"
				h.drone.Forward(100)
			case "Uw":
				info = "Stopped Going Forward"
				h.drone.Hover()
			case "Ds":
				info = "Started Going Backward"
				h.drone.Backward(100)
			case "Us":
				info = "Stopped Going Backward"
				h.drone.Hover()
			case "Da":
				info = "Started Going Left"
				h.drone.Left(100)
			case "Ua":
				info = "Stopped Going Left"
				h.drone.Hover()
			case "Dd":
				info = "Started Going Right"
				h.drone.Right(100)
			case "Ud":
				info = "Stopped Going Up"
				h.drone.Hover()
			case "Dr":
				info = "Started Going Up"
				h.drone.Up(100)
			case "Ur":
				info = "Stopped Going Up"
				h.drone.Hover()
			case "Df":
				info = "Started Going Down"
				h.drone.Down(100)
			case "Uf":
				info = "Stopped Going Down"
				h.drone.Hover()
			case "Du":
				info = "Started Take Off"
				h.drone.TakeOff()
			case "Dl":
				info = "Started Land"
				h.drone.Land()
			case "Uh":
				info = "Home set"
				if err := h.drone.SetHome(); err != nil {
					logrus.Error(err)
				}
				home = vector.V3D{
					float64(fd.MVO.PositionX),
					float64(fd.MVO.PositionY),
					float64(fd.MVO.PositionZ),
				}
				homeYaw = fd.IMU.Yaw
			case "U0":
				h.autoFlyTo(home, home, homeYaw)
			case "Un":
				fd := h.drone.GetFlightData()
				id := h.flyMap.AddCheckpoint(
					float64(fd.MVO.PositionX),
					float64(fd.MVO.PositionY),
					float64(fd.MVO.PositionZ),
				)
				if lastCheckpoint != 0 {
					h.flyMap.LinkCheckpoint(lastCheckpoint, id)
				}
				lastCheckpoint = id
			case "U1":
				h.autoFlyTo(h.flyMap.GetCheckpoint(1), home, homeYaw)
			case "U2":
				h.autoFlyTo(h.flyMap.GetCheckpoint(2), home, homeYaw)
			case "U3":
				h.autoFlyTo(h.flyMap.GetCheckpoint(3), home, homeYaw)
			case "U4":
				h.autoFlyTo(h.flyMap.GetCheckpoint(4), home, homeYaw)
			case "U5":
				h.autoFlyTo(h.flyMap.GetCheckpoint(5), home, homeYaw)
			case "U6":
				h.autoFlyTo(h.flyMap.GetCheckpoint(6), home, homeYaw)
			case "U7":
				h.autoFlyTo(h.flyMap.GetCheckpoint(7), home, homeYaw)
			case "U8":
				h.autoFlyTo(h.flyMap.GetCheckpoint(8), home, homeYaw)
			case "U9":
				h.autoFlyTo(h.flyMap.GetCheckpoint(9), home, homeYaw)
			default:
				info = string(msg.Content)
			}

			info += fmt.Sprintf(" BatPrc: %d; LgtStr: %d", fd.BatteryPercentage, fd.LightStrength)
			if fd.BatteryLow {
				info += " BatteryLow"
			}
			if fd.BatteryCritical {
				info += " BatteryCritical"
			}
			if fd.DownVisualState {
				info += " DownVisualState"
			}
			if fd.ErrorState {
				info += " ErrorState"
			}

			h.wsClient.SendMessage(wsclient.Message{
				Type:    wsclient.MTLog,
				Content: []byte("Command " + info),
			})
		}
	}
}

func (h *Controller) autoFlyTo(p vector.V3D, home vector.V3D, homeYaw int16) {
	p = p.Sub(home)
	h.wsClient.SendMessage(wsclient.Message{
		Type:    wsclient.MTLog,
		Content: []byte("Going home XY"),
	})
	doneXY, err := h.drone.AutoFlyToXY(float32(p.X()), float32(p.Y()))
	if err != nil {
		logrus.Error(err)
		return
	}
	go func() {
		<-doneXY
		h.wsClient.SendMessage(wsclient.Message{
			Type:    wsclient.MTLog,
			Content: []byte("Autoflight to XY done, Going home Yaw"),
		})
		doneYaw, err := h.drone.AutoTurnToYaw(homeYaw)
		if err != nil {
			logrus.Error(err)
			return
		}
		go func() {
			<-doneYaw
			h.wsClient.SendMessage(wsclient.Message{
				Type:    wsclient.MTLog,
				Content: []byte("Autoflight to home Yaw done, Going home Z"),
			})
			doneZ, err := h.drone.AutoFlyToHeight(int16(p.Z() / 10.))
			if err != nil {
				logrus.Error(err)
			}
			go func() {
				<-doneZ
				h.wsClient.SendMessage(wsclient.Message{
					Type:    wsclient.MTLog,
					Content: []byte("Autoflight to Z done"),
				})
			}()
		}()
	}()
}
