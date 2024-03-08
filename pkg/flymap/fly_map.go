package flymap

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/einherij/pilot/pkg/vector"
)

type FlyMap struct {
	mux         sync.RWMutex
	Name        string
	MtlLib      string
	checkpoints map[int]*Checkpoint
}

type Checkpoint struct {
	ID       int // autofilled
	Position vector.V3D
	Next     []*Checkpoint
}

func New(name, mtlLib string) *FlyMap {
	return &FlyMap{
		Name:        name,
		MtlLib:      mtlLib,
		checkpoints: make(map[int]*Checkpoint),
	}
}

func (fm *FlyMap) AddCheckpoint(x, y, z float64) (id int) {
	fm.mux.Lock()
	defer fm.mux.Unlock()

	for id = 1; ; id++ {
		if _, ok := fm.checkpoints[id]; !ok {
			fm.checkpoints[id] = &Checkpoint{
				ID:       id,
				Position: vector.V3D{x, y, z},
			}
			break
		}
	}
	return id
}

func (fm *FlyMap) GetCheckpoint(id int) vector.V3D {
	fm.mux.Lock()
	defer fm.mux.Unlock()

	return fm.checkpoints[id].Position
}

func (fm *FlyMap) LinkCheckpoint(fromID, toID int) {
	fm.mux.Lock()
	defer fm.mux.Unlock()

	from, ok := fm.checkpoints[fromID]
	if !ok {
		logrus.Warnf("checkpoints %d isn't found", fromID)
		return
	}
	to, ok := fm.checkpoints[toID]
	if !ok {
		logrus.Warnf("checkpoints %d isn't found", toID)
		return
	}
	from.Next = append(from.Next, to)
	to.Next = append(to.Next, from)
}

func LoadMap(path string) (*FlyMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer func() { _ = f.Close() }()
	m, err := ReadMap(f)
	if err != nil {
		return nil, fmt.Errorf("error reading map: %w", err)
	}
	return m, nil
}

func ReadMap(src io.Reader) (*FlyMap, error) {
	var atof = func(a string) float64 {
		f, err := strconv.ParseFloat(a, 64)
		if err != nil {
			logrus.Warnf("error converting string to float: %q", a)
		}
		return f
	}
	var atoi = func(a string) int {
		i, err := strconv.Atoi(a)
		if err != nil {
			logrus.Warnf("error converting string to int: %q", a)
		}
		return i
	}
	var m = &FlyMap{
		checkpoints: make(map[int]*Checkpoint),
	}
	br := bufio.NewReader(src)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading line: %w", err)
		}
		lineArr := strings.Split(string(line), " ")
		if len(lineArr) < 1 {
			logrus.Warnf("empty line")
			continue
		}
		switch lineArr[0] {
		case "o":
			if len(lineArr) < 2 {
				logrus.Warnf("broken name line")
				continue
			}
			m.Name = lineArr[1]
		case "mtllib":
			if len(lineArr) < 2 {
				logrus.Warnf("broken matlib line")
				continue
			}
			m.MtlLib = lineArr[1]
		case "v":
			if len(lineArr) < 4 {
				logrus.Warnf("broken matlib line")
				continue
			}
			m.AddCheckpoint(atof(lineArr[1]), atof(lineArr[2]), atof(lineArr[3]))
		case "l":
			if len(lineArr) < 3 {
				logrus.Warnf("broken matlib line")
				continue
			}
			m.LinkCheckpoint(atoi(lineArr[1]), atoi(lineArr[2]))
		case "":
		default:
			logrus.Warnf("unknown command")
			continue
		}
	}
	return m, nil
}

func SaveMap(path string, m *FlyMap) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()
	err = WriteMap(f, m)
	if err != nil {
		return fmt.Errorf("error writing map: %w", err)
	}
	return nil
}

func WriteMap(dest io.Writer, m *FlyMap) error {
	m.mux.RLock()
	defer m.mux.RUnlock()

	_, err := dest.Write([]byte("mtllib " + m.MtlLib + "\n"))
	if err != nil {
		return fmt.Errorf("error writing material library: %w", err)
	}
	_, err = dest.Write([]byte("o " + m.Name + "\n"))
	if err != nil {
		return fmt.Errorf("error writing name: %w", err)
	}
	err = m.forEach(func(checkpoint *Checkpoint) error {
		_, err = dest.Write([]byte(fmt.Sprintf("v %f %f %f\n", checkpoint.Position[0], checkpoint.Position[1], checkpoint.Position[2])))
		if err != nil {
			return fmt.Errorf("error writing name: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	type key struct {
		from int
		to   int
	}
	var linked = make(map[key]struct{})
	return m.forEach(func(checkpoint *Checkpoint) error {
		for _, next := range checkpoint.Next {
			from, to := checkpoint.ID, next.ID
			if from > to {
				from, to = to, from
			}
			k := key{from: from, to: to}
			if _, ok := linked[k]; ok {
				continue // link already written
			} else {
				linked[k] = struct{}{}
				_, err = dest.Write([]byte(fmt.Sprintf("l %d %d\n", from, to)))
				if err != nil {
					return fmt.Errorf("error writing name: %w", err)
				}
			}
		}
		return nil
	})
}

func (fm *FlyMap) forEach(f func(checkpoint *Checkpoint) error) error {
	for i := 1; ; i++ {
		if checkpoint, ok := fm.checkpoints[i]; !ok {
			break
		} else {
			if err := f(checkpoint); err != nil {
				return err
			}
		}
	}
	return nil
}

func (fm *FlyMap) GetOBJ() []byte {
	fm.mux.RLock()
	defer fm.mux.RUnlock()

	var buf bytes.Buffer
	_ = WriteMap(&buf, fm)
	return buf.Bytes()
}
