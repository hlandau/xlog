package xlog

// A sink which dispatches to zero or more other sinks.
type MultiSink struct {
	sinks []Sink
}

func (ms *MultiSink) Add(sink Sink) {
	for _, s := range ms.sinks {
		if s == sink {
			return
		}
	}

	ms.sinks = append(ms.sinks, sink)
}

func (ms *MultiSink) Remove(sink Sink) {
	var newSinks []Sink
	for _, s := range ms.sinks {
		if s != sink {
			newSinks = append(newSinks, s)
		}
	}

	ms.sinks = newSinks
}

func (ms *MultiSink) ReceiveLocally(sev Severity, format string, params ...interface{}) {
	for _, s := range ms.sinks {
		s.ReceiveLocally(sev, format, params...)
	}
}

func (ms *MultiSink) ReceiveFromChild(sev Severity, format string, params ...interface{}) {
	for _, s := range ms.sinks {
		s.ReceiveFromChild(sev, format, params...)
	}
}
