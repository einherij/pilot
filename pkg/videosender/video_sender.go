package videosender

import (
	"bufio"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const StreamCamera = `
	ffmpeg 
		-f avfoundation
		-framerate 30
		-video_size 640x480
		-i 0:none
		-vcodec libx264
		-preset ultrafast
		-tune zerolatency
		-g 40
		-acodec aac
		-f dash
		-dash_segment_type mp4
		-seg_duration 0.1
		-use_template 1
		-http_persistent 1
		%sdrone/video/fs/feed
`
const StreamPipe = `
	ffmpeg
		-i pipe:0
		-vcodec libx264
		-preset ultrafast
		-tune zerolatency
		-g 40
		-acodec aac
		-vf scale=320:180
		-f dash
		-dash_segment_type mp4
		-seg_duration 0.1
		-use_template 1
		-http_persistent 1
		%sdrone/video/fs/feed
`

type Sender struct {
	debugLog     bool
	command      string
	sourceStream <-chan []byte
	destURL      string
}

func New(destURL string, sourceStream <-chan []byte, command string, debugLog bool) *Sender {
	return &Sender{
		debugLog:     debugLog,
		command:      command,
		sourceStream: sourceStream,
		destURL:      destURL,
	}
}

func (s *Sender) Run(ctx context.Context) {
	logrus.Warnf("started video sender")
	const retryInterval = 5 * time.Second
	timer := time.NewTimer(0)
	name, args := parseCommand(s.command, s.destURL)

	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			logrus.Warnf("stopped video sender")
		case <-timer.C:
			func() {
				cmd := exec.CommandContext(ctx, name, args...)
				stdinPipe, err := cmd.StdinPipe()
				if err != nil {
					logrus.Error(fmt.Errorf("error opening stdin pipe: %w", err))
					return
				}
				defer func() { _ = stdinPipe.Close() }()

				stderrPipe, err := cmd.StderrPipe()
				if err != nil {
					logrus.Error(fmt.Errorf("error opening stderr pipe: %w", err))
					return
				}
				defer func() { _ = stderrPipe.Close() }()

				if err := cmd.Start(); err != nil {
					logrus.Error(fmt.Errorf("error starting command: %w", err))
				}

				if s.debugLog {
					go logStderr(ctx, stderrPipe)
				}

				if s.command == StreamPipe {
					for block := range s.sourceStream {
						if _, err := stdinPipe.Write(block); err != nil {
							logrus.Error(fmt.Errorf("error writing stream block: %w", err))
						}
					}
				}
			}()
			timer.Reset(retryInterval)
		}
	}
}

var spaceRegexp = regexp.MustCompile(`[\t\n\s]+`)

func parseCommand(command string, destURL string) (name string, args []string) {
	command = fmt.Sprintf(command, destURL)                     // replace destination URL
	command = spaceRegexp.ReplaceAllLiteralString(command, " ") // delete all tabs and new lines
	command = strings.TrimSpace(command)                        // delete left and right space
	lines := strings.Split(command, " ")
	if len(lines) < 2 {
		return
	}
	return lines[0], lines[1:]
}

func logStderr(ctx context.Context, stderrPipe io.ReadCloser) {
	bufStderr := bufio.NewReader(stderrPipe)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, _, err := bufStderr.ReadLine()
			if err != nil {
				logrus.Error(fmt.Errorf("error reading stderr: %w", err))
				return
			}
			if len(line) > 0 {
				logrus.WithField("label", "FFMPEG_STDERR").Warn(string(line))
			}
		}
	}
}
