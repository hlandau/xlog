package xlog

import "os"
import "io"
import "time"
import "fmt"

// Sink which writes each log message on a line to an io.Writer.
type WriterSink struct {
	w io.Writer
}

func NewWriterSink(w io.Writer) *WriterSink {
	return &WriterSink{w}
}

func (ws *WriterSink) ReceiveLocally(sev Severity, format string, params ...interface{}) {
	ws.ReceiveFromChild(sev, format, params...)
}

func (ws *WriterSink) ReceiveFromChild(sev Severity, format string, params ...interface{}) {
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
