package logging

import (
	// Logging

	"fmt"
	"os"
	"path/filepath"
	"runtime"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var (
	log      *logrus.Logger
	loglevel logrus.Level = logrus.TraceLevel
)

func init() {

	log = logrus.New()

	log.SetFormatter(&nested.Formatter{
		HideKeys: false,
		//FieldsOrder: []string{"component", "category"},
		//TimestampFormat: "2006-01-02 15:04:06", - DOES NOT WORK - does not update properly!
		//https://golang.org/pkg/time/#pkg-constants
		//TimestampFormat: "2006-01-02T15:04:05.999999999Z07:00",
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
		ShowFullLevel:   true,
	})

	// log to stdout not stderr
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above. - default
	log.SetLevel(loglevel)

}

// (re-)Define all the functions we want to use from logrus
// so that we can call them

// Info ...
func Info(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// Warn ...
func Warn(format string, v ...interface{}) {
	_, path, num, ok := runtime.Caller(1)
	if ok {
		file := filepath.Base(path)
		format = fmt.Sprintf("\033[1;35m%s#%d\033[0m, %s", file, num, format)
	}
	log.Warnf(format, v...)
}

// Error ...
func Error(format string, v ...interface{}) {
	_, path, num, ok := runtime.Caller(1)
	if ok {
		file := filepath.Base(path)
		format = fmt.Sprintf("\033[1;35m%s#%d\033[0m, %s", file, num, format)
	}
	log.Errorf(format, v...)
}

// Debug ...
func Debug(format string, v ...interface{}) {
	_, path, num, ok := runtime.Caller(1)
	if ok {
		file := filepath.Base(path)
		format = fmt.Sprintf("\033[1;35m%s#%d\033[0m, %s", file, num, format)
	}
	log.Debugf(format, v...)
}

// Fatal ...
func Fatal(format string, v ...interface{}) {
	_, path, num, ok := runtime.Caller(1)
	if ok {
		file := filepath.Base(path)
		format = fmt.Sprintf("\033[1;35m%s#%d\033[0m, %s", file, num, format)
	}
	log.Fatalf(format, v...)
}

// Panic ...
func Panic(format string, v ...interface{}) {
	_, path, num, ok := runtime.Caller(1)
	if ok {
		file := filepath.Base(path)
		format = fmt.Sprintf("\033[1;35m%s#%d\033[0m, %s", file, num, format)
	}
	log.Panicf(format, v...)
}

// Trace ...
func Trace(format string, v ...interface{}) {
	_, path, num, ok := runtime.Caller(1)
	if ok {
		file := filepath.Base(path)
		format = fmt.Sprintf("\033[1;35m%s#%d\033[0m, %s", file, num, format)
	}
	log.Tracef(format, v...)
}

// SetLevel ...
func SetLevel(loglevel logrus.Level) {
	log.SetLevel(loglevel)
}
