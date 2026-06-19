package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&nested.Formatter{
		HideKeys:        false,
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
		ShowFullLevel:   true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.TraceLevel)
}

// withCaller prepends "file#line, " to the format string so every level reports
// the source location consistently.
func withCaller(format string) string {
	if _, file, line, ok := runtime.Caller(2); ok {
		return fmt.Sprintf("\x1b[1;35m%s#%d\x1b[0m, %s", filepath.Base(file), line, format)
	}
	return format
}

func Info(format string, v ...interface{})  { log.Infof(withCaller(format), v...) }
func Warn(format string, v ...interface{})  { log.Warnf(withCaller(format), v...) }
func Error(format string, v ...interface{}) { log.Errorf(withCaller(format), v...) }
func Debug(format string, v ...interface{}) { log.Debugf(withCaller(format), v...) }
func Trace(format string, v ...interface{}) { log.Tracef(withCaller(format), v...) }
func Fatal(format string, v ...interface{}) { log.Fatalf(withCaller(format), v...) }
func Panic(format string, v ...interface{}) { log.Panicf(withCaller(format), v...) }

func SetLevel(level logrus.Level) { log.SetLevel(level) }
