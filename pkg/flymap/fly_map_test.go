package flymap

import (
	"bytes"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MapSuite struct {
	suite.Suite
}

func TestMapSuite(t *testing.T) {
	suite.Run(t, new(MapSuite))
}

func (s *MapSuite) TestReadMap() {
	obj := `mtllib map.mtl
o FlyMap
v 1.380002 -0.439160 -1.000000
v 1.380002 2.341999 -1.000000
v 1.000000 1.000000 1.000000
v 0.481311 1.000000 -1.000000
v -1.000000 1.000000 -1.000000
v 0.481311 2.341999 -1.000000
v -1.000000 1.000000 1.000000
v -1.017853 -1.000411 0.975697
v -1.085676 -0.439160 -1.000000
l 1 2
l 1 9
l 1 3
l 2 6
l 3 7
l 4 6
l 4 5
l 5 7
l 7 8
l 8 9
`
	m, err := ReadMap(bytes.NewBufferString(obj))
	s.NoError(err)

	var buf bytes.Buffer
	err = WriteMap(&buf, m)
	s.NoError(err)
	s.Equal(obj, buf.String())
}
