package xlog

import "io"
import "os"
import "fmt"
import "time"
import "sync"

type Severity int

const (
	SevNone      Severity = -1
	SevEmergency Severity = iota
	SevAlert
	SevCritical
	SevError
	SevWarn
	SevNotice
	SevInfo
	SevDebug
	SevTrace
)

var severityString = map[Severity]string{
	SevEmergency: "EMERGENCY", // EM EMR EMER
	SevAlert:     "ALERT",     // AL ALR ALER
	SevCritical:  "CRITICAL",  // CR CRT CRIT
	SevError:     "ERROR",     // ER ERR ERRO
	SevWarn:      "WARN",      // WA WRN WARN
	SevNotice:    "NOTICE",    // NO NOT NOTC
	SevInfo:      "INFO",      // IN INF INFO
	SevDebug:     "DEBUG",     // DE DBG DEBG
	SevTrace:     "TRACE",     // TR TRC TRAC
}

var severityValue = map[string]Severity{}

func init() {
	for k, v := range severityString {
		severityValue[v] = k
	}
}

func SeverityToString(severity Severity) string {
	return severityString[severity]
}

func ParseSeverity(severity string) (Severity, bool) {
	s, ok := severityValue[severity]
	return s, ok
}

// Site is the interface exposed to the externals of a package, which uses it
// to configure the logger. This is the other side of a Logger.
type Site interface {
	Name() string
	SetSeverity(severity Severity)
	SetSink(sink Sink)
}

// Logger is the interface exposed to the internals of a package, which uses it
// to log messages. This is the other side of a Site.
type Logger interface {
	Tracef(format string, params ...interface{})
	Debugf(format string, params ...interface{})
	Infof(format string, params ...interface{})
	Noticef(format string, params ...interface{})
	Warnf(format string, params ...interface{})
	Errorf(format string, params ...interface{})
	Criticalf(format string, params ...interface{})
	Alertf(format string, params ...interface{})
	Emergencyf(format string, params ...interface{})
	Sink
}

var loggersMutex sync.Mutex
var loggers = map[string]*logger{}

// Creates a new logger.
//
// Typical usage:
//
//     var log, Log = xlog.New("package name")
//
func New(name string) (Logger, Site) {
	loggersMutex.Lock()
	defer loggersMutex.Unlock()

	if _, ok := loggers[name]; ok {
		panic(fmt.Sprintf("Logger name conflict: logger with name %s already exists", name))
	}

	log := &logger{
		parent:      rootLogger,
		maxSeverity: SevDebug,
	}

	loggers[name] = log

	return log, log
}

type logger struct {
	maxSeverity Severity
	name        string

	parent Sink
}

type Sink interface {
	ReceiveFromChild(sev Severity, format string, params ...interface{})
}

var rootLogger = &logger{
	parent:      &writerSink{os.Stderr},
	maxSeverity: SevTrace,
}

var Root Site = rootLogger

func (l *logger) Name() string {
	return l.name
}

func (l *logger) SetSeverity(sev Severity) {
	l.maxSeverity = sev
}

func (l *logger) SetSink(sink Sink) {
	l.parent = sink
}

func (l *logger) Tracef(format string, params ...interface{}) {
	l.logf(SevTrace, format, params...)
}

func (l *logger) Debugf(format string, params ...interface{}) {
	l.logf(SevDebug, format, params...)
}

func (l *logger) Infof(format string, params ...interface{}) {
	l.logf(SevInfo, format, params...)
}

func (l *logger) Noticef(format string, params ...interface{}) {
	l.logf(SevNotice, format, params...)
}

func (l *logger) Warnf(format string, params ...interface{}) {
	l.logf(SevWarn, format, params...)
}

func (l *logger) Errorf(format string, params ...interface{}) {
	l.logf(SevError, format, params...)
}

func (l *logger) Criticalf(format string, params ...interface{}) {
	l.logf(SevCritical, format, params...)
}

func (l *logger) Alertf(format string, params ...interface{}) {
	l.logf(SevAlert, format, params...)
}

func (l *logger) Emergencyf(format string, params ...interface{}) {
	l.logf(SevEmergency, format, params...)
}

func (l *logger) logf(sev Severity, format string, params ...interface{}) {
	format = l.localPrefix() + format // XXX unsafe format string
	l.remoteLogf(sev, format, params...)
}

func (l *logger) remoteLogf(sev Severity, format string, params ...interface{}) {
	if sev > l.maxSeverity {
		return
	}

	if l.parent != nil {
		l.parent.ReceiveFromChild(sev, format, params...)
	}
}

func (l *logger) ReceiveFromChild(sev Severity, format string, params ...interface{}) {
	l.remoteLogf(sev, format, params...)
}

func (l *logger) localPrefix() string {
	if l.name != "" {
		return l.name + ": "
	}
	return ""
}

type writerSink struct {
	w io.Writer
}

func (ws *writerSink) ReceiveFromChild(sev Severity, format string, params ...interface{}) {
	msg := ws.prefix(sev) + fmt.Sprintf(format, params...) + "\n"
	io.WriteString(ws.w, msg)
}

func (ws *writerSink) prefix(sev Severity) string {
	return fmt.Sprintf("%s [%s] ", time.Now().Format("20060102150405"), severityString[sev])
}

type syslogger interface {
	Alert(m string) error
	Crit(m string) error
	Debug(m string) error
	Emerg(m string) error
	Err(m string) error
	Info(m string) error
	Notice(m string) error
	Warning(m string) error
}

type syslogSink struct {
	s syslogger
}

func (ss *syslogSink) ReceiveFromChild(sev Severity, format string, params ...interface{}) {
	s := fmt.Sprintf(format, params...)
	switch sev {
	case SevEmergency:
		ss.s.Emerg(s)
	case SevAlert:
		ss.s.Alert(s)
	case SevCritical:
		ss.s.Crit(s)
	case SevError:
		ss.s.Err(s)
	case SevWarn:
		ss.s.Warning(s)
	case SevNotice:
		ss.s.Notice(s)
	case SevInfo:
		ss.s.Info(s)
	default:
		ss.s.Debug(s)
	}
}

type LogClosure func() string

func (c LogClosure) String() string {
	return c()
}

func VisitSites(siteFunc func(s Site) error) error {
	loggersMutex.Lock()
	defer loggersMutex.Unlock()

	for _, v := range loggers {
		err := siteFunc(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetSeverityOnAllLoggers(severity Severity) {
	VisitSites(func(s Site) error {
		s.SetSeverity(severity)
		return nil
	})
}

func Flush() {
	// TODO
}
