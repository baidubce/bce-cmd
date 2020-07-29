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
	"testing"
)

type MatchType struct {
	pattern string
	name    string
	matched bool
}

func TestPatternFilter(t *testing.T) {

	testCases := []MatchType{
		// 1
		MatchType{
			pattern: "abs/*.jpg",
			name:    "abs/adbc/liupeng.jpg",
			matched: true,
		},
		// 2
		MatchType{
			pattern: "./abs/*/liupeng.jpg",
			name:    "abs/adbc/cdd/efg/liupeng.jpg",
			matched: false,
		},
		// 3
		MatchType{
			pattern: "abs/*/test/*.jpg",
			name:    "abs/adbc/dfdf/13213-3/test/liupeng.jpg",
			matched: true,
		},
		// 4
		MatchType{
			pattern: "abs/*/test/*.jpg",
			name:    "/abs/adbc/dfdf/13213-3/test/liupeng.jpg",
			matched: false,
		},
		// 5
		MatchType{
			pattern: "abs/*/test/*.jpg",
			name:    "./abs/adbc/dfdf/13213-3/test/liupeng.jpg",
			matched: false,
		},
		// 6
		MatchType{
			pattern: "~/abs/*.jpg",
			name:    "~/abs/adbc/liupeng.jpg",
			matched: true,
		},
		// 7
		MatchType{
			pattern: "*.jpg",
			name:    "abs/adbc/liupeng.jpg",
			matched: true,
		},
		// 8
		MatchType{
			pattern: "./abs/*.jpg",
			name:    "./abs/adbc/liupeng.png",
			matched: false,
		},
		// 9
		MatchType{
			pattern: "abs/*.jpg",
			name:    "abs/adbc/liupeng.g",
			matched: false,
		},
		// 10
		MatchType{
			pattern: "abs/[a-z].jpg",
			name:    "abs/1.jpg",
			matched: false,
		},
		// 11
		MatchType{
			pattern: "abs/[a-z].jpg",
			name:    "abs/g.jpg",
			matched: true,
		},
		// 12
		MatchType{
			pattern: "abs/liupeng-??.jpg",
			name:    "abs/liupeng-bj.jpg",
			matched: true,
		},
		// 13
		MatchType{
			pattern: "abs/liupeng-??.jpg",
			name:    "abs/liupeng-gz.jpg",
			matched: true,
		},
		// 14
		MatchType{
			pattern: "abs/liupeng/",
			name:    "abs/liupeng/gz.jpg",
			matched: false,
		},
	}

	for i, tCase := range testCases {

		matched, err := Match(tCase.pattern, tCase.name)
		ExpectEqual("match.go Match I", i+1, t.Errorf, true, err == nil)
		if err != nil {
			t.Logf("error: %s", err.Error())
			continue
		}
		ExpectEqual("filter_strategy PatternFilter  II", i+1, t.Errorf, tCase.matched,
			matched)
	}
}
