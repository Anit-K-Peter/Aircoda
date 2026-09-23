package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"aircoda/internal/config"
)

var (
	fileLogger *log.Logger
	mu         sync.Mutex
	initialized bool
)

// Init initializes logging to file.
func Init() {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		return
	}

	logPath := config.ResolveLogPath()
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err == nil {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			fileLogger = log.New(f, "[AIRCODA] ", log.LstdFlags)
		}
	}
	initialized = true
}

// Info logs informational messages to file.
func Info(format string, v ...interface{}) {
	Init()
	msg := fmt.Sprintf(format, v...)
	if fileLogger != nil {
		fileLogger.Println("INFO: " + msg)
	}
}

// Error logs error messages to file.
func Error(format string, v ...interface{}) {
	Init()
	msg := fmt.Sprintf(format, v...)
	if fileLogger != nil {
		fileLogger.Println("ERROR: " + msg)
	}
}

// PrintUserError displays a human-readable clean error message to stdout/stderr.
func PrintUserError(err error) {
	if err == nil {
		return
	}
	Error("User facing error: %v", err)
	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
}

// PrintUserErrorMsg displays a human-readable error message string.
func PrintUserErrorMsg(msg string) {
	Error("User facing error: %s", msg)
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}
