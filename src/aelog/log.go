/*
  Copyright 2012 Sergey Cherepanov (https://github.com/cheggaaa)

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package aelog

import (
	"log"
	"os"
	"cnst"
)



type AntLog struct {
	logger *log.Logger
	level  int
}


func New(filename string, level int) (*AntLog, error) {
	out := os.Stdout

	if filename != "" {
		var err error
		out, err = os.OpenFile(filename, os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
	}
	
	if level == 0 {
		level = cnst.LOG_INFO
	}	
	return &AntLog{log.New(out, "", log.LstdFlags), level}, nil
}


func (a *AntLog) Print(level int, v ...interface{}) {
	if level >= a.level {
		a.logger.Print(v...)
	}
}

func (a *AntLog) Printf(level int, format string, v ...interface{}) {
	if level >= a.level {
		a.logger.Printf(format, v...)
	}
}

func (a *AntLog) Println(level int, v ...interface{}) {
	if level >= a.level {
		a.logger.Println(v...)
	}
}

func (a *AntLog) Debug(v ...interface{}) {
	a.Print(cnst.LOG_DEBUG, v...)
}

func (a *AntLog) Info(v ...interface{}) {
	a.Print(cnst.LOG_INFO, v...)
}

func (a *AntLog) Warn(v ...interface{}) {
	a.Print(cnst.LOG_WARN, v...)
}

func (a *AntLog) Debugln(v ...interface{}) {
	a.Println(cnst.LOG_DEBUG, v...)
}

func (a *AntLog) Infoln(v ...interface{}) {
	a.Println(cnst.LOG_INFO, v...)
}

func (a *AntLog) Warnln(v ...interface{}) {
	a.Println(cnst.LOG_WARN, v...)
}

func (a *AntLog) Debugf(format string, v ...interface{}) {
	a.Printf(cnst.LOG_DEBUG, format, v...)
}

func (a *AntLog) Infof(format string, v ...interface{}) {
	a.Printf(cnst.LOG_INFO, format, v...)
}

func (a *AntLog) Warnf(format string, v ...interface{}) {
	a.Printf(cnst.LOG_WARN, format, v...)
}

func (a *AntLog) Fatal(v ...interface{}) {
	a.Warnln(v ...)
	log.Fatal(v ...)
}
