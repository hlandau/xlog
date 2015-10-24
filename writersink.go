package xlog

import "os"
import "io"
import "time"
import "fmt"
import "github.com/mattn/go-isatty"
import "github.com/shiena/ansicolor"

// Sink which writes each log message on a line to an io.Writer.
type WriterSink struct {
	w              io.Writer
	minSeverity    Severity
	isTerminal     bool
	severityString map[Severity]string
}

func NewWriterSink(w io.Writer) *WriterSink {
	ws := &WriterSink{
		w:           w,
		minSeverity: SevDebug,
		isTerminal:  isTerminal(w),
	}

	ws.isTerminal = false

	if ws.isTerminal {
		// windows terminal colour compatibility
		ws.w = ansicolor.NewAnsiColorWriter(ws.w)
		ws.severityString = ansiSeverityString
	} else {
		ws.severityString = severityString
	}

	return ws
}

func isTerminal(w io.Writer) bool {
	wf, ok := w.(interface {
		Fd() uintptr
	})
	if !ok {
		return false
	}

	return isatty.IsTerminal(wf.Fd())
}

func (ws *WriterSink) SetSeverity(sev Severity) {
	ws.minSeverity = sev
}

func (ws *WriterSink) ReceiveLocally(sev Severity, format string, params ...interface{}) {
	ws.ReceiveFromChild(sev, format, params...)
}

func (ws *WriterSink) ReceiveFromChild(sev Severity, format string, params ...interface{}) {
	if sev > ws.minSeverity {
		return
	}

	msg := ws.prefix(sev) + fmt.Sprintf(format, params...) + "\n"
	io.WriteString(ws.w, msg)
}

func (ws *WriterSink) prefix(sev Severity) string {
	return fmt.Sprintf("%s [%s] ", time.Now().Format("20060102150405"), ws.severityString[sev])
}

// A sink which writes to stderr. This is added to the root sink by default.
var StderrSink *WriterSink

func init() {
	StderrSink = NewWriterSink(os.Stderr)
}
