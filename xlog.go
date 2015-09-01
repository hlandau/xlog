package xlog

import "fmt"
import "sync"

// Site is the interface exposed to the externals of a package, which uses it
// to configure the logger. This is the other side of a Logger.
type Site interface {
	Name() string
	SetSeverity(severity Severity)
	SetSink(sink Sink)
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
		maxSeverity: SevTrace,
    name:        name,
	}

	loggers[name] = log

	return Logger{log}, log
}

type logger struct {
	maxSeverity Severity
	name        string

	parent Sink
}

type Sink interface {
	ReceiveLocally(sev Severity, format string, params ...interface{})
	ReceiveFromChild(sev Severity, format string, params ...interface{})
}

func init() {
	RootSink.Add(StderrSink)
}

var rootLogger = &logger{
	parent:      &RootSink,
	maxSeverity: SevTrace,
}

var Root Site = rootLogger

// The sink which is used by default by the root logger.
var RootSink MultiSink

func (l *logger) Name() string {
	return l.name
}

func (l *logger) SetSeverity(sev Severity) {
	l.maxSeverity = sev
}

func (l *logger) SetSink(sink Sink) {
	l.parent = sink
}

func (l *logger) ReceiveLocally(sev Severity, format string, params ...interface{}) {
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

// Calls a function for every Site which has been created.
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

// LogClosure can be used to pass a function that returns a string
// to a log method call. This is useful if the computation of a log message
// is expensive and the message will often be filtered.
type LogClosure func() string

func (c LogClosure) String() string {
	return c()
}
