package utils

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// SetupLogger configures and returns a logger instance
func SetupLogger(logLevel string) *log.Logger {
	logger := log.New()

	// Set log level
	switch logLevel {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}

	// Set custom formatter with colors and timestamps
	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
		DisableColors:   false,
		PadLevelText:    true,
	})

	// Output to stdout
	logger.SetOutput(os.Stdout)

	return logger
}

// SetupFileLogger configures logging to a file
func SetupFileLogger(logLevel, logFile string) (*log.Logger, error) {
	logger := SetupLogger(logLevel)

	if logFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}

		// Open log file
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}

		logger.SetOutput(file)
		
		// Disable colors for file output
		logger.SetFormatter(&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   true,
			PadLevelText:    true,
		})
	}

	return logger, nil
}