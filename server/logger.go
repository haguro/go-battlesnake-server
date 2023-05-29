package server

import "log"

const (
	LError = 1 << iota
	LWarning
	LInfo
	LDebug
	LDefault = LError | LWarning | LInfo
)

type Logger struct {
	*log.Logger
	opts int
}

func NewLogger(logger *log.Logger, opts int) *Logger {
	return &Logger{logger, opts}
}

func (l *Logger) Enabled(opt int) bool {
	return l.opts&opt != 0
}

func (l *Logger) Info(format string, v ...any) {
	if l.opts&LInfo != 0 {
		l.Printf("INFO "+format, v...)
	}
}

func (l *Logger) Warn(format string, v ...any) {
	if l.opts&LWarning != 0 {
		l.Printf("WARN "+format, v...)
	}
}
func (l *Logger) Err(format string, v ...any) {
	if l.opts&LError != 0 {
		l.Printf("ERROR "+format, v...)
	}
}

func (l *Logger) Debug(format string, v ...any) {
	if l.opts&LDebug != 0 {
		l.Printf("DEBUG "+format, v...)
	}
}
