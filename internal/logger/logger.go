package logger

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"time"
)

var Log *logrus.Logger

func Init() {
	Log = logrus.New()
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
	Log.SetOutput(os.Stdout)

	logDir := "logs"
	logFile := filepath.Join(logDir, "app.log")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		Log.Warnf("Не удалось создать директорию %s: %v", logDir, err)
		return
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Log.Warnf("Не удалось открыть лог-файл %s: %v", logFile, err)
		return
	}

	Log.SetOutput(file)
	Log.Info("Лог-файл подключён:", logFile)
}
