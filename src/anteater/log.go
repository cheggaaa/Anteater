package anteater

import (
	"log"
	"os"
)

const (
	LOG_DEBUG = 0
	LOG_INFO  = 1
	LOG_WARN  = 2
)

var Log *AntLog

type AntLog struct {
	logger *log.Logger
	level  int
}


func LogInit() error {
	out := os.Stderr
	
	if len(Conf.LogFile) > 0 {
		var err error
		out, err = os.OpenFile(Conf.LogFile, os.O_WRONLY | os.O_APPEND, 0644)
		if err != nil {
			return err
		}
	}
	
	Log = &AntLog{log.New(out, "", log.LstdFlags), Conf.LogLevel}
	return nil
}


func (a *AntLog) Print(level int, v ...interface{}) {
	if a.level >= level {
		a.logger.Print(v...)
	}
}

func (a *AntLog) Printf(level int, format string, v ...interface{}) {
	if a.level >= level {
		a.logger.Printf(format, v...)
	}
}

func (a *AntLog) Println(level int, v ...interface{}) {
	if a.level >= level {
		a.logger.Println(v...)
	}
}

func (a *AntLog) Debug(v ...interface{}) {
	a.Print(LOG_DEBUG, v...)
}

func (a *AntLog) Info(v ...interface{}) {
	a.Print(LOG_INFO, v...)
}

func (a *AntLog) Warn(v ...interface{}) {
	a.Print(LOG_WARN, v...)
}

func (a *AntLog) Debugln(v ...interface{}) {
	a.Println(LOG_DEBUG, v...)
}

func (a *AntLog) Infoln(v ...interface{}) {
	a.Println(LOG_INFO, v...)
}

func (a *AntLog) Warnln(v ...interface{}) {
	a.Println(LOG_WARN, v...)
}

func (a *AntLog) Debugf(format string, v ...interface{}) {
	a.Printf(LOG_DEBUG, format, v...)
}

func (a *AntLog) Infof(format string, v ...interface{}) {
	a.Printf(LOG_INFO, format, v...)
}

func (a *AntLog) Warnf(format string, v ...interface{}) {
	a.Printf(LOG_WARN, format, v...)
}

func (a *AntLog) Fatal(v ...interface{}) {
	a.Warnln(v ...)
	log.Fatal(v ...)
}
