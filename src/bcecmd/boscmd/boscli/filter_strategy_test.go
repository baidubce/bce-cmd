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

package boscli

import (
	// 	"fmt"
	// 	"os"
	// 	"runtime"
	"strings"
	"testing"
	// 	"time"
	"path/filepath"
)

import (
	"utils/util"
)

type getAbsPatternType struct {
	orgPattern string
	retPattern string
	isSuc      bool
}

func TestGetAbsPattern(t *testing.T) {

	thisDir, err := filepath.Abs("./")
	if err != nil {
		t.Errorf("Get current path failed! error: %s", err)
		return
	}
	homeDir, err := util.GetHomeDirOfUser()
	if err != nil {
		t.Errorf("Get user home dir failed! error: %s", err)
		return
	}
	pathLists := strings.Split(thisDir, util.OsPathSeparator)

	testCases := []getAbsPatternType{
		//1
		getAbsPatternType{
			orgPattern: "",
			retPattern: thisDir,
			isSuc:      true,
		},
		//2
		getAbsPatternType{
			orgPattern: "abs/*.jpg",
			retPattern: filepath.Join(thisDir, "abs/*.jpg"),
			isSuc:      true,
		},
		//3
		getAbsPatternType{
			orgPattern: "./abc/enfg/*",
			retPattern: filepath.Join(thisDir, "abc/enfg/*"),
			isSuc:      true,
		},
		//4
		getAbsPatternType{
			orgPattern: "../*",
			retPattern: filepath.Join(util.GetParentPath(pathLists, 1), "*"),
			isSuc:      true,
		},
		//5
		getAbsPatternType{
			orgPattern: "../../*.jpg",
			retPattern: filepath.Join(util.GetParentPath(pathLists, 2), "*.jpg"),
			isSuc:      true,
		},
		//6
		getAbsPatternType{
			orgPattern: "../../",
			retPattern: util.GetParentPath(pathLists, 2),
			isSuc:      true,
		},
		//7
		getAbsPatternType{
			orgPattern: "~/*",
			retPattern: filepath.Join(homeDir, "*"),
			isSuc:      true,
		},
		//8
		getAbsPatternType{
			orgPattern: "~/*.svn",
			retPattern: filepath.Join(homeDir, "*.svn"),
			isSuc:      true,
		},
		//9
		getAbsPatternType{
			orgPattern: "~/test/*",
			retPattern: filepath.Join(homeDir, "test/*"),
			isSuc:      true,
		},
	}

	for i, tCase := range testCases {
		ret, err := getAbsPattern(tCase.orgPattern)
		util.ExpectEqual("filter_strategy getAbsPattern I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if tCase.isSuc && err == nil {
			util.ExpectEqual("filter_strategy getAbsPattern  II", i+1, t.Errorf, tCase.retPattern,
				ret)
		} else if err != nil {
			t.Logf("error: %s", err.Error())
		}
	}
}

type setPatternsType struct {
	exclude     []string
	include     []string
	retPatterns []string
	isLocal     bool
	isInclude   bool
	code        BosCliErrorCode
}

func TestSetPatterns(t *testing.T) {

	thisDir, err := filepath.Abs("./")
	if err != nil {
		t.Errorf("Get current path failed! error: %s", err)
		return
	}
	homeDir, err := util.GetHomeDirOfUser()
	if err != nil {
		t.Errorf("Get user home dir failed! error: %s", err)
		return
	}

	testCases := []setPatternsType{
		// 1 exclude local
		setPatternsType{
			exclude: []string{"abs/*.jpg", "./abc/enfg/*"},
			retPatterns: []string{
				filepath.Join(thisDir, "abs/*.jpg"),
				filepath.Join(thisDir, "abc/enfg/*"),
			},
			isLocal: true,
			code:    BOSCLI_OK,
		},
		// 2 exclude bos
		setPatternsType{
			exclude: []string{"bos:/abs/*.jpg", "bos:/abc/enfg/*"},
			retPatterns: []string{
				"abs/*.jpg",
				"abc/enfg/*",
			},
			code: BOSCLI_OK,
		},
		// 2 exclude bos
		setPatternsType{
			exclude: []string{"bos://abs/*.jpg", "bos://abc/enfg/*"},
			retPatterns: []string{
				"abs/*.jpg",
				"abc/enfg/*",
			},
			code: BOSCLI_OK,
		},
		// 3 exclude local
		setPatternsType{
			exclude:     []string{"~/abs/*.jpg"},
			retPatterns: []string{filepath.Join(homeDir, "abs/*.jpg")},
			isLocal:     true,
			code:        BOSCLI_OK,
		},
		// 4 have exclude include
		setPatternsType{
			exclude: []string{"abs/*.jpg"},
			include: []string{"abs/*.jpg"},
			code:    BOSCLI_SYNC_EXCLUDE_INCLUDE_TOG,
		},
		// 5 exclude bos
		setPatternsType{
			exclude:     []string{"bos:/abs/*.jpg"},
			retPatterns: []string{"abs/*.jpg"},
			code:        BOSCLI_OK,
		},
		// 5 exclude bos
		setPatternsType{
			exclude:     []string{"bos://abs/*.jpg"},
			retPatterns: []string{"abs/*.jpg"},
			code:        BOSCLI_OK,
		},

		// 6 include local
		setPatternsType{
			include: []string{"abs/*.jpg", "./abc/enfg/*"},
			retPatterns: []string{
				filepath.Join(thisDir, "abs/*.jpg"),
				filepath.Join(thisDir, "abc/enfg/*"),
			},
			isLocal:   true,
			isInclude: true,
			code:      BOSCLI_OK,
		},
		// 7 include local
		setPatternsType{
			include:     []string{"abs/*.jpg"},
			retPatterns: []string{filepath.Join(thisDir, "abs/*.jpg")},
			isLocal:     true,
			isInclude:   true,
			code:        BOSCLI_OK,
		},
		// 8 include bos
		setPatternsType{
			include: []string{"bos:/abs/*.jpg", "bos:/abc/enfg/*"},
			retPatterns: []string{
				"abs/*.jpg",
				"abc/enfg/*",
			},
			isInclude: true,
			code:      BOSCLI_OK,
		},
		// 9 include local
		setPatternsType{
			include:     []string{"abs/*.jpg"},
			retPatterns: []string{"abs/*.jpg"},
			isInclude:   true,
			code:        BOSCLI_OK,
		},
	}

	for i, tCase := range testCases {
		filter := &bosFilter{}
		retCode, err := filter.setPatterns(tCase.exclude, tCase.include, tCase.isLocal)
		util.ExpectEqual("filter_strategy setPatterns I", i+1, t.Errorf, tCase.code, retCode)
		if err == nil {
			util.ExpectEqual("filter_strategy setPatterns  II", i+1, t.Errorf, tCase.retPatterns,
				filter.patterns)
		} else if err != nil {
			t.Logf("error: %s", err.Error())
		}
	}
}

type patternFilterType struct {
	exclude     []string
	include     []string
	excludeTime []string
	includeTime []string
	bosPath     string
	isLocal     bool
	filtered    bool
}

func TestPatternFilter(t *testing.T) {

	testCases := []patternFilterType{
		// 1 exclude local
		patternFilterType{
			exclude:  []string{"abs/*.jpg", "./abc/enfg/*"},
			bosPath:  "abs/adbc/liupeng.jpg",
			isLocal:  true,
			filtered: true,
		},
		// 2 exclude local
		patternFilterType{
			exclude:  []string{"abs/*.jpg", "./abc/enfg/*"},
			bosPath:  "./abs/adbc/liupeng.jpg",
			isLocal:  true,
			filtered: true,
		},
		// 3 exclude bos
		patternFilterType{
			exclude:  []string{"bos:/abs/*.jpg", "bos:/abc/enfg/*"},
			bosPath:  "bos:/abs/adbc/liupeng.jpg",
			filtered: true,
		},
		// 4 exclude bos
		patternFilterType{
			exclude:  []string{"bos://abs/*.jpg", "bos://abc/enfg/*"},
			bosPath:  "bos:/abs/adbc/liupeng.jpg",
			filtered: true,
		},
		// 5 exclude bos
		patternFilterType{
			exclude:  []string{"bos://abs/*.jpg", "bos://abc/enfg/*"},
			bosPath:  "bos://abs/adbc/liupeng.jpg",
			filtered: true,
		},
		// 6 exclude local
		patternFilterType{
			exclude:  []string{"~/abs/*.jpg"},
			bosPath:  "~/abs/adbc/liupeng.jpg",
			filtered: true,
			isLocal:  true,
		},
		// 7 exclude bos
		patternFilterType{
			exclude:  []string{"bos:/abs/*.jpg"},
			bosPath:  "abs/adbc/liupeng.jpg",
			filtered: true,
		},
		// 8 exclude bos
		patternFilterType{
			exclude:  []string{"bos:/abs/*.jpg"},
			bosPath:  "abs/adbc/liupeng.png",
			filtered: false,
		},
		// 9 exclude bos
		patternFilterType{
			exclude:  []string{"bos:/abs/*.jpg"},
			bosPath:  "abs/adbc/liupeng.png",
			filtered: false,
		},
		// 10 exclude bos
		patternFilterType{
			exclude:  []string{"bos:/abs/*.jpg"},
			bosPath:  "bos:/abs/adbc/liupeng.png",
			filtered: false,
		},
		// 11 exclude bos
		patternFilterType{
			exclude:  []string{"./abs/*.jpg"},
			bosPath:  "./abs/adbc/liupeng.png",
			filtered: false,
		},
		// 12 include local
		patternFilterType{
			include:  []string{"abs/*.jpg", "./abc/enfg/*"},
			bosPath:  "./abs/adbc/liupeng.jpg",
			isLocal:  true,
			filtered: false,
		},
		// 13 include local
		patternFilterType{
			include:  []string{"abs/*.jpg", "./abc/enfg/*"},
			bosPath:  "abs/adbc/liupeng.jpg",
			isLocal:  true,
			filtered: false,
		},
		// 14 include local
		patternFilterType{
			include:  []string{"abs/*.jpg", "./abc/enfg/*"},
			bosPath:  "abs/adbc/liupeng.g",
			isLocal:  true,
			filtered: true,
		},
		// 15 include local
		patternFilterType{
			include:  []string{"abs/*.jpg", "./abc/enfg/*"},
			bosPath:  "./abc/enfg/liupeng.g",
			isLocal:  true,
			filtered: false,
		},
		// 16 include local
		patternFilterType{
			include:  []string{"bos:/abs/*.jpg", "bos:/abc/enfg/*"},
			bosPath:  "bos:/abc/enfg/liupeng.g",
			filtered: false,
		},
		// 17 include local
		patternFilterType{
			include:  []string{"bos:/abs/*.jpg", "bos:/abc/enfg/*"},
			bosPath:  "bos:/abs/liupeng.g",
			filtered: true,
		},
		// 18 include local
		patternFilterType{
			include:  []string{"bos:/abc/enfg"},
			bosPath:  "bos:/abc/enfg/",
			filtered: false,
		},
		// 19 exclude bos
		patternFilterType{
			exclude:  []string{"bos:/abs/*.jpg"},
			bosPath:  "abs/adbc/liupeng.png/",
			filtered: false,
		},
		// 20 exclude
		patternFilterType{
			exclude:  []string{"./abs/*/liupeng.jpg"},
			bosPath:  "abs/adbc/cdd/efg/liupeng.jpg",
			isLocal:  true,
			filtered: true,
		},
	}

	for i, tCase := range testCases {
		filter, retCode, err := newSyncFilter(tCase.exclude, tCase.include, tCase.excludeTime,
			tCase.includeTime, tCase.isLocal)

		if retCode != BOSCLI_OK {
			t.Errorf("ID: %d, new filter failed, error: %v", i+1, err)
			continue
		}

		bosPath := tCase.bosPath
		if tCase.isLocal {
			if bosPath, err = util.Abs(tCase.bosPath); err != nil {
				t.Errorf("ID: %d, get abs failed, error: %v", i+1, err)
				continue
			}
		}

		filtered, err := filter.PatternFilter(bosPath)
		util.ExpectEqual("filter_strategy PatternFilter I", i+1, t.Errorf, true, err == nil)
		if err != nil {
			t.Logf("error: %s", err.Error())
			continue
		}
		util.ExpectEqual("filter_strategy PatternFilter  II", i+1, t.Errorf, tCase.filtered,
			filtered)

		if len(tCase.exclude) > 0 {
			filtered, err = filter.ExcludePatternFilter(bosPath)

			util.ExpectEqual("filter_strategy PatternFilter III", i+1, t.Errorf, true, err == nil)
			if err != nil {
				t.Logf("error: %s", err.Error())
				continue
			}
			util.ExpectEqual("filter_strategy PatternFilter  IV", i+1, t.Errorf, tCase.filtered,
				filtered)
		}
	}
}
