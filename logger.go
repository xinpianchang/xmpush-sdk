package xmpush

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
}

type nopeLogger struct{}

func (l *nopeLogger) Debug(args ...interface{}) {
	// nope
}

func (l *nopeLogger) Debugf(format string, args ...interface{}) {
	// nope
}

type simpleLogger struct {
	log *log.Logger
}

func newSimpleLogger() *simpleLogger {
	return &simpleLogger{
		log: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *simpleLogger) fileInfo() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

func (l *simpleLogger) Debug(args ...interface{}) {
	l.log.Println("[debug]", l.fileInfo(), fmt.Sprint(args...))
}

func (l *simpleLogger) Debugf(format string, args ...interface{}) {
	l.log.Println("[debug]", l.fileInfo(), fmt.Sprintf(format, args...))
}
