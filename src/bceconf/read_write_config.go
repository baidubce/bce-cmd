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

// bcecg.go: loading configure from file and write cofigure to file.

package bceconf

import (
	"fmt"
	"os"
	"reflect"
	"sync"
)

import (
	"code.google.com/p/gcfg"
)

var (
	rwmutex sync.RWMutex
)

/*
 *Loading configure from file.
 *     Use gcfg loading configure from file.
 *PARAMS:
 *     confPath: the path of configure file.
 *     cfg     : the config struct
 *RETURN:
 *     error:  nil or error.
 */
func LoadConfig(confPath string, cfg interface{}) error {
	rwmutex.RLock()
	defer rwmutex.RUnlock()
	if err := gcfg.ReadFileInto(cfg, confPath); err != nil {
		return err
	}
	return nil
}

func writeStructToCofingFile(fd *os.File, cfgStructType reflect.Type,
	cfgStructValue reflect.Value) error {
	for i := 0; i < cfgStructValue.NumField(); i++ {
		t := cfgStructValue.Field(i)
		f := cfgStructType.Field(i)
		fmt.Fprintf(fd, "%s = %v\n", f.Name, t.Interface())
	}
	fmt.Fprintf(fd, "\n")
	return nil
}

/*
 * Writing configure into File.
 * PARAMS:
 *		confPath: the path of configure file.
 *      cfg: a interface, which must be a point of struct, and the fields of it must be struct
 *      or map, where map must have string keys and pointer-to-struct values!
 * RETURN:
 *      error: nil or error.
 */
func WriteConfig(confPath string, cfg interface{}) error {
	// TODO need lock conf file ?
	fd, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer fd.Close()

	cfgValue := reflect.ValueOf(cfg)
	elemOfCfgValue := cfgValue.Elem()
	typeOfCfg := elemOfCfgValue.Type()

	if cfgValue.Kind() != reflect.Ptr || elemOfCfgValue.Kind() != reflect.Struct {
		return fmt.Errorf("cofing must be a point of struct")
	}

	numberOfField := elemOfCfgValue.NumField()
	for i := 0; i < numberOfField; i++ {
		cField := elemOfCfgValue.Field(i)
		cType := typeOfCfg.Field(i)
		if cField.Kind() == reflect.Struct {
			fmt.Fprintf(fd, "[%s]\n", cType.Name)
			writeStructToCofingFile(fd, cType.Type, cField)
		} else if cField.Kind() == reflect.Map {
			mapType := cField.Type()
			if mapType.Key().Kind() != reflect.String || mapType.Elem().Kind() != reflect.Ptr ||
				mapType.Elem().Elem().Kind() != reflect.Struct {
				return fmt.Errorf("section must have string keys and pointer-to-struct values!")
			}
			for _, k := range cField.MapKeys() {
				mapValue := cField.MapIndex(k)
				fmt.Fprintf(fd, "[%s \"%s\"]\n", cType.Name, k.Interface())
				writeStructToCofingFile(fd, mapValue.Elem().Type(), mapValue.Elem())
			}
		} else {
			return fmt.Errorf("only support struct and map")
		}
	}
	return nil
}
