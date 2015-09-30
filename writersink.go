package xlog

import "os"
import "io"
import "time"
import "fmt"

// Sink which writes each log message on a line to an io.Writer.
type WriterSink struct {
	w           io.Writer
	minSeverity Severity
}

func NewWriterSink(w io.Writer) *WriterSink {
	return &WriterSink{
		w:           w,
		minSeverity: SevDebug,
	}
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
	return fmt.Sprintf("%s [%s] ", time.Now().Format("20060102150405"), severityString[sev])
}

// A sink which writes to stderr. This is added to the root sink by default.
var StderrSink *WriterSink

func init() {
	StderrSink = NewWriterSink(os.Stderr)
}
