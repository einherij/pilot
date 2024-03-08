package tellointer

import (
	"time"

	"github.com/SMerrony/tello"
)

type Drone interface {
	ControlConnectDefault() (err error)
	ControlDisconnect()

	VideoConnectDefault() (<-chan []byte, error)
	VideoDisconnect()
	SetVideoWide()
	GetVideoSpsPps()

	StreamFlightData(asAvailable bool, periodMs time.Duration) (<-chan tello.FlightData, error)

	TakeOff()
	Land()
	Hover()
	Forward(pct int)
	Backward(pct int)
	Left(pct int)
	Right(pct int)
	Up(pct int)
	Down(pct int)
	TurnRight(pct int)
	TurnLeft(pct int)
}
