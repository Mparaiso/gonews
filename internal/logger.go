// @copyrights 2016 mparaiso <mparaiso@online.fr>

package gonews

import (
	"log"
	"os"
	"runtime"
	"time"
)

// LoggerInterface defines a logger
type LoggerInterface interface {
	Debug(messages ...interface{})
	Info(messages ...interface{})
	Error(messages ...interface{})
	ErrorWithStack(messages ...interface{})
}

const time_format = "2006-01-02 15:04:05"

// LogLevel , @see http://www.tutorialspoint.com/log4j/log4j_logging_levels.htm
type LogLevel int8

const (
	ALL LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

// Logger is a logger
type Logger struct {
	*log.Logger
	level LogLevel
}

// NewDefaultLogger returns a logger using the default log package
func NewDefaultLogger(level LogLevel) *Logger {
	return &Logger{log.New(os.Stdout, "", 0), level}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug logs a debugging message
func (l *Logger) Debug(messages ...interface{}) {
	if l.level <= DEBUG {
		l.Logger.Print(append([]interface{}{"\r[DEBUG] ", time.Now().Format(time_format), "\n\t"}, messages...)...)
	}
}

// Info logs an info message
func (l *Logger) Info(messages ...interface{}) {
	if l.level <= INFO {
		l.Logger.Print(append([]interface{}{"\r[INFO] ", time.Now().Format(time_format), "\n\t"}, messages...)...)
	}
}

// Error logs an error message
func (l *Logger) Error(messages ...interface{}) {
	if l.level <= ERROR {
		l.Logger.Print(append([]interface{}{"\r[ERROR] ", time.Now().Format(time_format), "\n\t"}, messages...)...)
	}

}

// ErrorWithStack displays a stack trace
func (l *Logger) ErrorWithStack(messages ...interface{}) {
	if l.level <= ERROR {
		l.Error(messages...)
		// @see http://stackoverflow.com/questions/19094099/how-to-dump-goroutine-stacktraces/19712747#19712747
		buffer := make([]byte, 1<<16)
		runtime.Stack(buffer, false)
		//print 6 lines max
		l.Logger.Printf("\r[ERROR] %s \r\t%s", time.Now().Format(time_format), buffer)
	}
}
