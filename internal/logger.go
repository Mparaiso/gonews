// @copyrights 2016 mparaiso <mparaiso@online.fr>

package gonews

import (
	"log"
	"os"
	"runtime"
	"time"
)

const time_format = "2006-01-02 15:04:05"

// LoggerInterface defines a logger
type LoggerInterface interface {
	Debug(messages ...interface{})
	Info(messages ...interface{})
	Error(messages ...interface{})
	ErrorWithStack(messages ...interface{})
}

// Logger is a logger
type Logger struct {
	*log.Logger
	IsDebug bool
}

// NewDefaultLogger returns a logger using the default log package
func NewDefaultLogger(debug bool) *Logger {
	return &Logger{log.New(os.Stdout, "", 0), debug}
}

// Debug logs a debugging message
func (l *Logger) Debug(messages ...interface{}) {
	if l.IsDebug {
		l.Logger.Print(append([]interface{}{"[DEBUG] ", time.Now().Format(time_format), "\n\t"}, messages...)...)
	}
}

// Info logs an info message
func (l *Logger) Info(messages ...interface{}) {
	l.Logger.Print(append([]interface{}{"[INFO] ", time.Now().Format(time_format), "\n\t"}, messages...)...)
}

// Error logs an error message
func (l *Logger) Error(messages ...interface{}) {
	l.Logger.Print(append([]interface{}{"[ERROR] ", time.Now().Format(time_format), "\n\t"}, messages...)...)

}

// ErrorWithStack displays a stack trace
func (l *Logger) ErrorWithStack(messages ...interface{}) {
	l.Error(messages...)
	// @see http://stackoverflow.com/questions/19094099/how-to-dump-goroutine-stacktraces/19712747#19712747
	buffer := make([]byte, 1<<16)
	runtime.Stack(buffer, false)
	//print 6 lines max
	l.Logger.Printf("\r%s", buffer)
}
