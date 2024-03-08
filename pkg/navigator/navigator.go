package navigator

import (
	"context"
	"fmt"
	"github.com/SMerrony/tello"
	"github.com/einherij/pilot/pkg/vector"
	"github.com/sirupsen/logrus"
	"sync/atomic"
)

type Position struct {
	Location vector.V3D
	Rotation vector.V3D
}

type Navigator struct {
	flightData <-chan tello.FlightData
	currentPos atomic.Pointer[Position] // Position
}

func NewNavigator(flightData <-chan tello.FlightData) *Navigator {
	n := &Navigator{
		flightData: flightData,
	}
	n.currentPos.Store(&Position{
		Location: vector.V3D{0, 0, 0},
		Rotation: vector.V3D{1, 0, 0},
	})
	return n
}

func (n *Navigator) Run(ctx context.Context) {
	logrus.Warnf("started navigation")
	for {
		select {
		case fd := <-n.flightData:
			currentPos := *(n.currentPos.Load())
			currentPos.Location = vector.V3D{
				float64(fd.MVO.PositionX),
				float64(fd.MVO.PositionY),
				float64(-fd.MVO.PositionZ),
			}
			singleVector := vector.V3D{1., 0., 0.}
			currentPos.Rotation = singleVector.RotateZ(float64(fd.IMU.Yaw))
			n.currentPos.Store(&currentPos)
		case <-ctx.Done():
			logrus.Warnf("stopped navigation")
			return
		}
	}
}

func (n *Navigator) GetPos() Position {
	pos := *(n.currentPos.Load())
	return pos
}

func (p Position) GetOBJ() []byte {
	dirLeft := p.Rotation.RotateZ(-135).Add(p.Location)
	dirRight := p.Rotation.RotateZ(135).Add(p.Location)
	directionLocation := p.Rotation.Add(p.Location)
	return []byte(
		fmt.Sprintf(
			"mtllib pos.mtl\n"+
				"o Pos\n"+
				"v %f %f %f\n"+
				"v %f %f %f\n"+
				"v %f %f %f\n"+
				"v %f %f %f\n"+
				"l 1 2\n"+
				"l 2 3\n"+
				"l 2 4\n"+
				"l 4 1\n"+
				"l 3 1\n"+
				"f 2/1/2 4/1/4 1/1/1 3/1/3\n",
			p.Location[0], p.Location[1], p.Location[2],
			directionLocation[0], directionLocation[1], directionLocation[2],
			dirLeft[0], dirLeft[1], dirLeft[2],
			dirRight[0], dirRight[1], dirRight[2],
		))
}
