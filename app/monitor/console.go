package monitor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chzyer/readline"
)

type consoleWriter struct {
	rl        *readline.Instance
	w         io.Writer
	line      bytes.Buffer
	backlog   bytes.Buffer
	timer     *time.Timer
	mutex     sync.Mutex
	interval  time.Duration // minimum time between display updates
	maxUpdate int           // maximum number of charaters per update
}

func newConsoleWriter(rl *readline.Instance) io.Writer {
	cw := &consoleWriter{
		rl:        rl,
		w:         os.Stdout,
		interval:  time.Millisecond * 100,
		maxUpdate: 2000,
	}
	return cw
}

func (c *consoleWriter) Write(p []byte) (int, error) {
	for _, b := range p {
		c.line.WriteByte(b)
		if b == '\n' {
			c.backlog.Write(c.line.Bytes())
			if c.timer == nil {
				c.timer = time.AfterFunc(c.interval, c.emit)
			}
			c.line.Reset()
		}
	}
	return len(p), nil
}

func (c *consoleWriter) emit() {
	c.mutex.Lock()
	defer func() {
		c.mutex.Unlock()
	}()

	if c.backlog.Len() == 0 {
		c.timer = nil
		return
	}
	update := c.backlog.String()
	lines := strings.Count(update, "\n")
	omission := false
	if lines > c.maxUpdate {
		// Count backwards to find start of the first line in the
		// maximum lines allowed per update
		omission = true
		seen := 0
		for i := len(update) - 1; i >= 0; i-- {
			if update[i] == '\n' {
				seen++
				if seen == c.maxUpdate {
					update = update[i+1 : len(update)-1]
					break
				}
			}
		}
	}
	// carriage return to go to the begnning of the line
	// then ansi escape sequence to clear the line
	c.w.Write([]byte("\r\033[2K"))
	if omission {
		text := fmt.Sprintf("... omitted %v lines\n", lines-c.maxUpdate)
		c.w.Write([]byte(text))
	}
	c.w.Write([]byte(update))
	c.rl.Refresh()
	c.backlog.Reset()
	c.timer = time.AfterFunc(c.interval, c.emit)
}
