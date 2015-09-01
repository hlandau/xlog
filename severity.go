package xlog

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

func (severity Severity) String() string {
	return severityString[severity]
}

func ParseSeverity(severity string) (Severity, bool) {
	s, ok := severityValue[severity]
	return s, ok
}
