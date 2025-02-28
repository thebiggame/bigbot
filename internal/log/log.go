package log

import (
	"fmt"
	"log"
	"os"
)

var Logger *log.Logger

// Initialise default logger.
func init() {
	Logger = log.New(os.Stdout, "", log.LstdFlags)
}

func Trace(v ...interface{}) {
	fa := "[TRACE] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func Tracef(format string, v ...interface{}) {
	Trace(fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	fa := "[DEBUG] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func Debugf(format string, v ...interface{}) {
	Debug(fmt.Sprintf(format, v...))
}

func Warn(v ...interface{}) {
	fa := "[WARN] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func Warnf(format string, v ...interface{}) {
	Warn(fmt.Sprintf(format, v...))
}

func Info(v ...interface{}) {
	fa := "[INFO] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func Infof(format string, v ...interface{}) {
	Info(fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	fa := "[ERROR] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func Errorf(format string, v ...interface{}) {
	Error(fmt.Sprintf(format, v...))
}

func Fatal(v ...interface{}) {
	fa := "[FATAL] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func Fatalf(format string, v ...interface{}) {
	Fatal(fmt.Sprintf(format, v...))
}
