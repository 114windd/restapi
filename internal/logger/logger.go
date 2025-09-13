package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init() {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)

	// Set format based on environment
	if os.Getenv("ENV") == "production" {
		Log.SetFormatter(&logrus.JSONFormatter{})
		Log.SetLevel(logrus.InfoLevel)
	} else {
		Log.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
		Log.SetLevel(logrus.DebugLevel)
	}
}

// Helper functions for common logging patterns
func LogRequest(method, path, userID string) *logrus.Entry {
	return Log.WithFields(logrus.Fields{
		"method":  method,
		"path":    path,
		"user_id": userID,
		"type":    "request",
	})
}

func LogDatabase(operation, table string) *logrus.Entry {
	return Log.WithFields(logrus.Fields{
		"operation": operation,
		"table":     table,
		"type":      "database",
	})
}

func LogAuth(action, email string) *logrus.Entry {
	return Log.WithFields(logrus.Fields{
		"action": action,
		"email":  email,
		"type":   "auth",
	})
}
