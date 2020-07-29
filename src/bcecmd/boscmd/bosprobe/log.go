// Copyright 2017 Baidu, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
// except in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the
// License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing permissions
// and limitations under the License.

package bosprobe

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	PROBE_LOG_FORMAT = "bosprobe%s_%d.log"
	PROBE_LOG_
)

type probeLog struct {
	logFd              *os.File
	logName            string
	createLogFileError bool
}

// Generate a new file for save logs
// if failed to create log file, print logs to stdout.
func (l *probeLog) createLogFile() {

	l.logName = fmt.Sprintf(PROBE_LOG_FORMAT, time.Now().Format("2006-01-02_15_04_05"),
		rand.Intn(100000))

	var err error
	if l.logFd, err = os.Create(l.logName); err != nil {
		l.logFd = os.Stdout
		l.logName = ""
		l.createLogFileError = true
	}
}

func (l *probeLog) logging(format string, args ...interface{}) {
	fmt.Fprintf(l.logFd, format, args...)
}

// save to log file and print to terminal
func (l *probeLog) Tlogging(format string, args ...interface{}) {
	fmt.Fprintf(l.logFd, format, args...)
	if !l.createLogFileError {
		fmt.Printf(format, args...)
	}
}
