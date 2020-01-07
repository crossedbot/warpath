package logger

import (
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

var log Logger
var once sync.Once
var Log = func() Logger {
	once.Do(func() {
		l := logrus.New()
		l.Formatter = &logrus.TextFormatter{DisableColors: true}
		l.SetLevel(logrus.InfoLevel)
		log = Logger{Logger: l}
	})
	return log
}()

func SetFile(path string) error {
	if path != "" {
		fd, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		mw := io.MultiWriter(os.Stderr, fd)
		Log.Out = mw
	}
	return nil
}

func Debug(args ...interface{}) {
	Log.Debug(args)
}

func Info(args ...interface{}) {
	Log.Info(args)
}

func Error(args ...interface{}) {
	Log.Error(args)
}

func Warning(args ...interface{}) {
	Log.Warning(args)
}

func Fatal(args ...interface{}) {
	Log.Fatal(args)
}
