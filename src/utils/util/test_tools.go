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

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

import (
	"github.com/baidubce/bce-sdk-go/bce"
)

func CreateFileWithSize(filePath string, fileSize int64) error {
	fd, err := os.Create(filePath)
	if err != nil {
		return err
	}
	return fd.Truncate(fileSize)
}

func InitTestFiles() {
	if !DoesDirExist("./test_file") {
		err := TryMkdir("./test_file")
		if err != nil {
			fmt.Printf("Errorf: mkdir ./test_file failed\n")
			os.Exit(2)
		}
	}
	if !DoesFileExist("./test_file/clod.rar") {
		if err := CreateFileWithSize("./test_file/clod.rar", 3e8); err != nil {
			fmt.Printf("Error: create ./test_file/clod.rar failed, error: %s\n", err.Error())
			os.Exit(2)
		}
	}
	if !DoesFileExist("./test_file/xcache.txt") {
		if err := CreateFileWithSize("./test_file/xcache.txt", 1e3); err != nil {
			fmt.Printf("Error: create ./test_file/xcache.txt failed, error: %s\n", err.Error())
			os.Exit(2)
		}
	}
}

// create a temp file and write content to this file
func CreateAnRandomFileWithContent(format string, args ...interface{}) (*os.File, string, error) {
	fileName := "cli_test_" + GetRandomString(10)
	fd, err := os.Create(fileName)
	if err != nil {
		return nil, "", fmt.Errorf("create file input_tmep filed")
	}
	fmt.Fprintf(fd, format, args...)
	fd.Close()
	fd, err = os.OpenFile(fileName, os.O_RDONLY, 0755)
	if err != nil {
		os.Remove(fileName)
		return nil, "", fmt.Errorf("open file input_temp filed")
	}
	return fd, fileName, nil
}

func ExpectEqual(funcName string, id int, alert func(format string, args ...interface{}), expected,
	actual interface{}) bool {
	if expected == nil && actual == nil {
		return true
	} else if expected == nil && actual != nil {
		alert("%s id: %d get %v want %v\n", funcName, id, actual, expected)
		return false
	} else if expected != nil && actual == nil {
		alert("%s id: %d get %v want %v\n", funcName, id, actual, expected)
		return false
	}
	expectedValue := reflect.ValueOf(expected)
	equal := false
	if actualType := reflect.TypeOf(actual); actualType != nil {
		if expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) {
			equal = reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
		}
	}
	if !equal {
		alert("%s id: %d get %v want %v\n", funcName, id, actual, expected)
		return false
	}
	return true
}

func ErrorEqual(funcName string, id int, alert func(format string, args ...interface{}), expected,
	actual interface{}) bool {
	if expected == nil && actual == nil {
		return true
	} else if expected == nil && actual != nil {
		alert("%s id: %d get %v want %v\n", funcName, id, actual, expected)
		return false
	} else if expected != nil && actual == nil {
		alert("%s id: %d get %v want %v\n", funcName, id, actual, expected)
		return false
	}
	exServerErr, exOk := expected.(*bce.BceServiceError)
	serverErr, ok := actual.(*bce.BceServiceError)

	if exOk && ok {
		if exServerErr.Code != serverErr.Code {
			alert("%s id: %d get %v want %v\n", funcName, id, serverErr.Code, exServerErr.Code)
			return false
		}
		return true
	} else if !exOk && !ok {
		// Do nothing
	} else {
		alert("%s id: %d get %v want %v\n", funcName, id, actual, expected)
		return false
	}
	return true
}

func GetParentPath(lists []string, upLevel int) string {
	if runtime.GOOS != "windows" {
		if len(lists) <= upLevel {
			return "/"
		}
		return OsPathSeparator + filepath.Join(lists[:len(lists)-upLevel]...)
	}
	if len(lists) < 1 || len(lists) <= upLevel+1 {
		return "/"
	}
	return strings.Join(lists[:len(lists)-upLevel], OsPathSeparator)
}
