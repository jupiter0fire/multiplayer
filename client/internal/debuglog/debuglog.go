/*
Package debuglog logs the messages that the client hands to the daemon.

The daemon broadcasts exactly these messages to its websocket subscribers, so a
record written here is identical to the payload that goes over the websocket.
*/
package debuglog

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
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
	return l != nil && (l.console || l.file != nil)
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
daemon; it is written verbatim so that the log file contains the exact websocket
payload.
*/
func (l *Logger) Log(msgType string, data []byte) {
	if !l.Enabled() {
		return
	}

	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		fmt.Fprintf(l.file, "%s %s\n", now.Format(time.RFC3339Nano), data)
	}
	if l.console {
		fmt.Fprintf(os.Stdout, "%s %s %s\n",
			color.HiBlackString(now.Format("15:04:05.000")),
			typeColor(msgType).Sprint(msgType),
			data,
		)
	}
}

/* Logf records a free-form diagnostic message, used for errors that occur while forwarding a message. */
func (l *Logger) Logf(format string, args ...any) {
	if !l.Enabled() {
		return
	}

	line := fmt.Sprintf(format, args...)
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		fmt.Fprintf(l.file, "%s ERROR %s\n", now.Format(time.RFC3339Nano), line)
	}
	if l.console {
		fmt.Fprintf(os.Stdout, "%s %s %s\n",
			color.HiBlackString(now.Format("15:04:05.000")),
			color.HiRedString("ERROR"),
			line,
		)
	}
}

/* typeColor returns the console color used for a message type. */
func typeColor(msgType string) *color.Color {
	switch msgType {
	case "GAME_START":
		return color.New(color.FgHiGreen)
	case "GAME_END":
		return color.New(color.FgHiRed)
	case "INFO_ITEM":
		return color.New(color.FgHiYellow)
	case "INFO_ENTRANCE":
		return color.New(color.FgHiCyan)
	default:
		return color.New(color.Reset)
	}
}
