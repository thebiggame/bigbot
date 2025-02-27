package log

import (
	"fmt"
	"log"
)

var Logger *log.Logger

// SetLogger sets the active logger to the Log-compliant logger passed.
func SetLogger(logger *log.Logger) {
	Logger = logger
}

func LogDebug(v ...interface{}) {
	fa := "[DEBUG] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func LogDebugf(format string, v ...interface{}) {
	LogDebug(fmt.Sprintf(format, v...))
}

func LogWarn(v ...interface{}) {
	fa := "[WARN] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func LogWarnf(format string, v ...interface{}) {
	LogWarn(fmt.Sprintf(format, v...))
}

func LogInfo(v ...interface{}) {
	fa := "[INFO] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func LogInfof(format string, v ...interface{}) {
	LogInfo(fmt.Sprintf(format, v...))
}

func LogErr(v ...interface{}) {
	fa := "[ERROR] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func LogErrf(format string, v ...interface{}) {
	LogErr(fmt.Sprintf(format, v...))
}

func LogFatal(v ...interface{}) {
	fa := "[FATAL] "
	v = append([]interface{}{fa}, v...)
	Logger.Print(v...)
}

func LogFatalf(format string, v ...interface{}) {
	LogFatal(fmt.Sprintf(format, v...))
}
