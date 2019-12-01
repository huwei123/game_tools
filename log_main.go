package main

import (
	"fmt"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

func init() {

	// Use the Airbrake hook to report errors that have Error severity or above to
	// an exception tracker. You can create custom hooks, see the Hooks section.
	//log.AddHook(airbrake.NewHook(123, "xyz", "production"))

	//logrus_syslog.

	//hook, err := logrus_syslog.NewSyslogHook("udp", "localhost:514", syslog.LOG_INFO, "")
	//if err != nil {
	//	log.Error("Unable to connect to local syslog daemon")
	//} else {
	//	log.AddHook(hook)
	//}
}

//func ConfigLocalFilesystemLogger(logPath string, logFileName string, maxAge time.Duration, rotationTime time.Duration) {
//	baseLogPath := path.Join(logPath, logFileName)
//	write, err := rotatelogs.New()
//}
//
//func newLfsHook(logLevel *string, maxRemainCnt uint) logrus.Hook {
//	rotatelogs.
//	write, err := rotatelogs.New(
//		logName+".%Y%m%d%H",
//	)
//}

func main() {
	logf, err := rotatelogs.New(
		"f:\\logs\\test_log.%Y%m%d%H%M",
		//rotatelogs.WithLinkName("/path/to/access_log"),
		rotatelogs.WithMaxAge(5*time.Minute),
		rotatelogs.WithRotationTime(time.Minute),
	)
	if err != nil {
		fmt.Println(err)
	}

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: logf, // 为不同级别设置不同的输出目的
		logrus.InfoLevel:  logf,
		logrus.WarnLevel:  logf,
		logrus.ErrorLevel: logf,
		logrus.FatalLevel: logf,
		logrus.PanicLevel: logf,
	}, &logrus.TextFormatter{DisableColors: true})

	logrus.AddHook(lfHook)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			logrus.WithFields(logrus.Fields{
				"animal": "walrus",
			}).Info("A walrus appears")
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()

	//logrus.SetFormatter(&logrus.TextFormatter{
	//	DisableColors: true,
	//	FullTimestamp: true,
	//})
	//logrus.SetFormatter(&logrus.JSONFormatter{})
	//
	//logrus.SetReportCaller(true)
	//
	//
	//logrus.WithFields(logrus.Fields{
	//	"animal": "walrus",
	//}).Info("A walrus appears")
	//
	//
	//logrus.Info("AERASERSERSER")
}
