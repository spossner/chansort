package utils

import (
	"log"
	"os"
)

var debugLogger *log.Logger

func init() {
	// Create debug log file
	file, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	debugLogger = log.New(file, "TUI: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Debug logs debug information to file
func Debug(format string, v ...interface{}) {
	if debugLogger != nil {
		debugLogger.Printf(format, v...)
	}
}
