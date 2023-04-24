package migration

import "log"

const (
	warningFlag = "⚠️⚠️⚠️, waring! "
	successFlag = "✅✅✅, success! "
	failedFlag  = "❌❌❌, failed! "
	logPrefix   = "[migration] "
)

type Logger struct {
	*log.Logger
}

func NewLogger(l *log.Logger) *Logger {
	if l == nil {
		l = log.Default()
	}
	return &Logger{l}
}

func (l *Logger) Info(v ...interface{}) {
	l.Logger.Println(append([]interface{}{logPrefix}, v...)...)
}

func (l *Logger) InfoWithFlag(err error, v ...interface{}) {
	if err != nil {
		l.Logger.Println(append(append([]interface{}{logPrefix, failedFlag}, v...), " ,Error:", err)...)
	} else {
		l.Logger.Println(append([]interface{}{logPrefix, successFlag}, v...)...)
	}
}

func (l *Logger) WarnWithFlag(v ...interface{}) {
	l.Logger.Println(append([]interface{}{logPrefix, warningFlag}, v...)...)
}
