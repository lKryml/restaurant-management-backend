package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Log *logrus.Logger

func InitLogger() {
	Log = logrus.New()
	Log.Out = os.Stdout

	Log.SetReportCaller(true)

	Log.SetLevel(logrus.DebugLevel)
	Log.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})
}
