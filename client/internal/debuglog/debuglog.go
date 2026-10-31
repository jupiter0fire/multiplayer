/*
Package debuglog logs the messages that the client hands to the daemon.

The daemon re-serializes each message before broadcasting it to its websocket
subscribers, so a record written here is semantically identical to the payload
that goes over the websocket, but not necessarily byte-for-byte identical after
a daemon upgrade. Only client-to-daemon messages are recorded: a record does not
prove that the daemon delivered the message, for example when a slow subscriber
falls behind and the daemon drops a broadcast.
*/
package debuglog

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	typeColors = map[string]*color.Color{
		"GAME_START":    color.New(color.FgHiGreen),
		"GAME_END":      color.New(color.FgHiRed),
		"INFO_ITEM":     color.New(color.FgHiYellow),
		"INFO_ENTRANCE": color.New(color.FgHiCyan),
		"ERROR":         color.New(color.FgHiRed),
	}
	defaultTypeColor = color.New(color.Reset)
)

/* Logger writes one timestamped record per message to the console and/or to a log file. */
type Logger struct {
	mu      sync.Mutex
	console bool
	file    *os.File
}

/* Open creates a logger. verbose enables the console output, path enables the log file. */
func Open(verbose bool, path string) (*Logger, error) {
	l := &Logger{console: verbose}
	if path == "" {
		return l, nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	l.file = f
	return l, nil
}

/* Enabled reports whether the logger produces any output. A nil logger writes nothing. */
func (l *Logger) Enabled() bool {
	if l == nil {
		return false
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabledLocked()
}

func (l *Logger) enabledLocked() bool {
	return l.console || l.file != nil
}

/* Close closes the log file, if one is open. */
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return nil
	}

	err := l.file.Close()
	l.file = nil
	return err
}

/*
Log records one message. data must be the serialized message as it is sent to the
daemon; it is written verbatim so that the log file contains the message payload.
*/
func (l *Logger) Log(msgType string, data []byte) {
	l.write(msgType, string(data))
}

/* Logf records a free-form diagnostic message, used for errors that occur while forwarding a message. */
func (l *Logger) Logf(format string, args ...any) {
	l.write("ERROR", singleLine(fmt.Sprintf(format, args...)))
}

func (l *Logger) write(msgType string, payload string) {
	if l == nil {
		return
	}

	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.enabledLocked() {
		return
	}

	if l.file != nil {
		fmt.Fprintf(l.file, "%s %s %s\n", now.Format(time.RFC3339Nano), msgType, payload)
	}
	if l.console {
		fmt.Fprintf(color.Output, "%s %s %s\n",
			color.HiBlackString(now.Format("15:04:05.000")),
			typeColor(msgType).Sprint(msgType),
			payload,
		)
	}
}

/* singleLine escapes line breaks so that one record stays on one line. */
func singleLine(s string) string {
	return strings.NewReplacer("\r", `\r`, "\n", `\n`).Replace(s)
}

/* typeColor returns the console color used for a message type. */
func typeColor(msgType string) *color.Color {
	if c, ok := typeColors[msgType]; ok {
		return c
	}
	return defaultTypeColor
}
