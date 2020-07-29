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

// This module provides the filter strategy for pattern exclude and time filter.

// time range [start time, end time]

package boscli

import (
	// 	"fmt"
	"strings"
)

import (
	"bcecmd/boscmd"
	"utils/util"
)

type timeRange struct {
	start int64
	end   int64
}

type bosFilter struct {
	pathFilterIsInclude bool
	timeFilterIsInclude bool
	pathFilterEnabled   bool
	timeFilterEnabled   bool
	patterns            []string
	timeRanges          []timeRange
	pathSeparator       string
}

func getAbsPattern(pattern string) (string, error) {
	if strings.HasPrefix(pattern, util.OsPathSeparator) || strings.HasPrefix(pattern, "*") {
		return pattern, nil
	}
	absPattern, err := util.Abs(pattern)
	if err != nil {
		return "", err
	}
	return absPattern, nil
}

func newSyncFilter(exclude, include, excludeTime, includeTime []string,
	srcIsLocal bool) (*bosFilter, BosCliErrorCode, error) {

	filter := &bosFilter{}
	if retCode, err := filter.setPatterns(exclude, include, srcIsLocal); retCode != BOSCLI_OK {
		return nil, retCode, err
	}

	if retCode, err := filter.setTimeRanges(excludeTime, includeTime); retCode != BOSCLI_OK {
		return nil, retCode, err
	}

	return filter, BOSCLI_OK, nil
}

// Set exclude or include
func (b *bosFilter) setPatterns(exclude, include []string, srcIsLocal bool) (BosCliErrorCode,
	error) {

	var (
		patternsTemp []string
	)

	if len(exclude) > 0 && len(include) > 0 {
		return BOSCLI_SYNC_EXCLUDE_INCLUDE_TOG, nil
	}

	if len(include) > 0 {
		patternsTemp = include
		b.pathFilterIsInclude = true
	} else {
		patternsTemp = exclude
	}

	if len(patternsTemp) > 0 {
		if srcIsLocal {
			b.pathSeparator = util.OsPathSeparator
			for _, pattern := range patternsTemp {
				// tran pattern to abs pattern
				absPattern, err := getAbsPattern(pattern)
				if err != nil {
					return BOSCLI_EMPTY_CODE, err
				}
				b.patterns = append(b.patterns, absPattern)
			}
		} else {
			b.pathSeparator = boscmd.BOS_PATH_SEPARATOR
			for _, pattern := range patternsTemp {
				// remove bos prefix
				pattern = FilterPrefixOfBosPath(pattern)
				b.patterns = append(b.patterns, pattern)
			}
		}
		b.pathFilterEnabled = true
	}
	return BOSCLI_OK, nil
}

// Set time ranges of time filter
func (b *bosFilter) setTimeRanges(excludeTime, includeTime []string) (BosCliErrorCode, error) {
	var (
		timeTemp []string
	)

	if len(excludeTime) > 0 && len(includeTime) > 0 {
		return BOSCLI_SYNC_EXCLUDE_INCLUDE_TIME_TOG, nil
	}

	// is exclude or include ?
	if len(includeTime) > 0 {
		timeTemp = includeTime
		b.timeFilterIsInclude = true
	} else {
		timeTemp = excludeTime
	}

	if len(timeTemp) > 0 {
		b.timeFilterEnabled = true
	}

	return BOSCLI_OK, nil
}

// Filter with path pattern
func (b *bosFilter) PatternFilter(bosPath string) (bool, error) {
	// if path have bos prefix, remove bos prefix

	bosPath = FilterPrefixOfBosPath(bosPath)
	for _, pattern := range b.patterns {
		if strings.HasSuffix(bosPath, b.pathSeparator) && !strings.HasSuffix(pattern,
			b.pathSeparator) {
			pattern += b.pathSeparator
		}
		matched, err := util.Match(pattern, bosPath)
		if err != nil {
			return false, err
		}
		if matched {
			if b.pathFilterIsInclude {
				return false, nil
			} else {
				return true, nil
			}
		}
	}

	if b.pathFilterIsInclude {
		return true, nil
	}
	return false, nil
}

// Only work for exclude local file with pattern
func (b *bosFilter) ExcludePatternFilter(bosPath string) (bool, error) {
	if b.pathFilterIsInclude {
		return false, nil
	}

	// if path have bos prefix, remove bos prefix
	bosPath = FilterPrefixOfBosPath(bosPath)

	for _, pattern := range b.patterns {
		if strings.HasSuffix(bosPath, b.pathSeparator) && !strings.HasSuffix(pattern,
			b.pathSeparator) {
			pattern += b.pathSeparator
		}
		matched, err := util.Match(pattern, bosPath)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// Time filter
func (b *bosFilter) TimeFilter(mtime int64) bool {
	for _, timeRange := range b.timeRanges {
		if mtime >= timeRange.start && mtime <= timeRange.end {
			if b.timeFilterIsInclude {
				return false
			}
			return true
		}
	}

	if b.timeFilterIsInclude {
		return true
	}
	return false
}
