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

// This module compare file between files.

package boscli

import (
	"fmt"
	"strings"
	"time"
)

import (
	"bcecmd/boscmd"
	"utils/util"
)

type existenceType string

const (
	FILE_AT_BOTH_SIDE existenceType = "fileAtBothSide"
	FILE_NOT_AT_SRC   existenceType = "fileNotAtSrc"
	FILE_NOT_AT_DST   existenceType = "fileNotAtDst"
)

type syncOpDetail struct {
	syncFunc    string
	srcFileInfo *fileDetail
	dstFileInfo *fileDetail
	srcPath     string
	dstPath     string
	err         error
	ended       bool
}

// generate new comparator
// PARAM:
//      atBothSide, sync strategy of both side have file
//      notAtDst, sync strategy of dst don't have file
//      notAtSrc, sync strategy of src don't have file
//      args, sync information
func NewComparator(atBothSide, notAtDst, notAtSrc syncStrategyInfterface,
	args *syncArgs, srcFilesIterator, dstFilesIterator fileListIterator) *Comparator {
	comp := &Comparator{
		fileAtBothSideSyncStrategy: atBothSide,
		fileNotAtSrcSyncStrategy:   notAtSrc,
		fileNotAtDstSyncStrategy:   notAtDst,
		syncOpChan:                 make(chan syncOpDetail, args.syncProcessingNum),
		syncInfo:                   args,
	}
	go comp.compare(srcFilesIterator, dstFilesIterator)
	return comp
}

type Comparator struct {
	fileAtBothSideSyncStrategy syncStrategyInfterface
	fileNotAtSrcSyncStrategy   syncStrategyInfterface
	fileNotAtDstSyncStrategy   syncStrategyInfterface
	syncOpChan                 chan syncOpDetail
	syncInfo                   *syncArgs
}

// An iterator of sync operation
func (c *Comparator) compare(srcFilesIterator fileListIterator, dstFilesIterator fileListIterator) {
	var (
		srcAllItered bool
		dstAllItered bool
		srcTakeNext  bool = true
		dstTakeNext  bool = true
		srcFileInfo  *fileDetail
		dstFileInfo  *fileDetail
	)

	for {
		// get dst next file detail (file maybe object or local file)
		if !srcAllItered && srcTakeNext {
			srcListResult, err := srcFilesIterator.next()
			if err != nil {
				c.syncOpChan <- syncOpDetail{err: err}
				break
			}

			if srcListResult.err != nil {
				c.syncOpChan <- syncOpDetail{err: srcListResult.err}
				break
			}

			if srcListResult.ended {
				srcAllItered = true
			} else {
				srcFileInfo = srcListResult.file
				if srcFileInfo.err != nil {
					c.syncOpChan <- syncOpDetail{err: srcFileInfo.err}
					// maybe this file is deleted before upload or download, so continue
					if ErrIsNotExist(srcFileInfo.err) {
						continue
					} else {
						break
					}
				}
			}
		}

		// get dst next file detail (file maybe object or local file)
		if !dstAllItered && dstTakeNext {

			dstListResult, err := dstFilesIterator.next()

			if err != nil {
				c.syncOpChan <- syncOpDetail{err: err}
				break
			}

			if dstListResult.err != nil {
				c.syncOpChan <- syncOpDetail{err: dstListResult.err}
				break
			}

			if dstListResult.ended {
				dstAllItered = true
			} else {
				dstFileInfo = dstListResult.file
				if dstFileInfo.err != nil {
					c.syncOpChan <- syncOpDetail{err: dstFileInfo.err}
					// maybe this file is deleted before upload or download, so continue
					if ErrIsNotExist(srcFileInfo.err) {
						srcTakeNext = false
						continue
					} else {
						break
					}
				}
			}
		}

		// compare dst and src, generate sync operations
		if !srcAllItered && !dstAllItered {
			fileExistence := c.compareFileExistence(srcFileInfo, dstFileInfo)
			if fileExistence == FILE_AT_BOTH_SIDE {
				srcTakeNext = true
				dstTakeNext = true
				needSync, err := c.fileAtBothSideSyncStrategy.shouldSync(srcFileInfo, dstFileInfo)
				if err != nil {
					c.syncOpChan <- syncOpDetail{err: err}
					break
				} else if needSync {
					c.syncOpChan <- syncOpDetail{
						syncFunc:    c.fileAtBothSideSyncStrategy.genSyncFunc(),
						srcPath:     srcFileInfo.path,
						dstPath:     dstFileInfo.path,
						srcFileInfo: srcFileInfo,
						dstFileInfo: dstFileInfo,
					}
				}
			} else if fileExistence == FILE_NOT_AT_DST {
				srcTakeNext = true
				dstTakeNext = false
				if c.fileNotAtDstSyncStrategy != nil {
					needSync, err := c.fileNotAtDstSyncStrategy.shouldSync(srcFileInfo, nil)
					if err != nil {
						c.syncOpChan <- syncOpDetail{err: err}
						break
					} else if needSync {
						c.syncOpChan <- syncOpDetail{
							syncFunc:    c.fileNotAtDstSyncStrategy.genSyncFunc(),
							srcPath:     srcFileInfo.path,
							dstPath:     c.deduceDstFullPath(srcFileInfo),
							srcFileInfo: srcFileInfo,
						}
					}
				}
			} else if fileExistence == FILE_NOT_AT_SRC {
				srcTakeNext = false
				dstTakeNext = true
				if c.fileNotAtSrcSyncStrategy != nil {
					needSync, err := c.fileNotAtSrcSyncStrategy.shouldSync(nil, dstFileInfo)
					if err != nil {
						c.syncOpChan <- syncOpDetail{err: err}
						break
					} else if needSync {
						c.syncOpChan <- syncOpDetail{
							syncFunc:    c.fileNotAtSrcSyncStrategy.genSyncFunc(),
							dstPath:     dstFileInfo.path,
							dstFileInfo: dstFileInfo,
						}
					}
				}
			}
		} else if srcAllItered && !dstAllItered {
			// if no 'not_at_src' strategy is not specified and all src iterated
			// no need to iterate all dst files

			if c.fileNotAtSrcSyncStrategy == nil {
				break
			}
			dstTakeNext = true
			needSync, err := c.fileNotAtSrcSyncStrategy.shouldSync(nil, dstFileInfo)
			if err != nil {
				c.syncOpChan <- syncOpDetail{err: err}
				break
			} else if needSync {
				c.syncOpChan <- syncOpDetail{
					syncFunc:    c.fileNotAtSrcSyncStrategy.genSyncFunc(),
					dstPath:     dstFileInfo.path,
					dstFileInfo: dstFileInfo,
				}
			}
		} else if !srcAllItered && dstAllItered {
			// if no 'not_at_dst' strategy is not specified and all dst iterated
			// no need to iterate all src files
			if c.fileNotAtDstSyncStrategy == nil {
				break
			}
			srcTakeNext = true
			needSync, err := c.fileNotAtDstSyncStrategy.shouldSync(srcFileInfo, nil)
			if err != nil {
				c.syncOpChan <- syncOpDetail{err: err}
				break
			} else if needSync {
				c.syncOpChan <- syncOpDetail{
					syncFunc:    c.fileNotAtDstSyncStrategy.genSyncFunc(),
					srcPath:     srcFileInfo.path,
					dstPath:     c.deduceDstFullPath(srcFileInfo),
					srcFileInfo: srcFileInfo,
				}
			}
		} else {

			break
		}
	}
	c.syncOpChan <- syncOpDetail{
		ended: true,
	}
}

// Get next sync operation.
func (c *Comparator) next() (*syncOpDetail, error) {
	select {
	case syncOpVal := <-c.syncOpChan:
		return &syncOpVal, nil
	case <-time.After(time.Duration(SYNC_COMPARATOR_TIME_OUT) * time.Millisecond):
		return nil, fmt.Errorf("Get next sync operation time out!")
	}
	return nil, fmt.Errorf("Unknown Error!")
}

// Compares src file and dst file to determine file existence
func (c *Comparator) compareFileExistence(src *fileDetail, dst *fileDetail) existenceType {
	if src.key == dst.key {
		return FILE_AT_BOTH_SIDE
	} else if src.key < dst.key {
		return FILE_NOT_AT_DST
	}
	return FILE_NOT_AT_SRC
}

// Deduces the full path of dst from src file sync info
func (c *Comparator) deduceDstFullPath(srcFileInfo *fileDetail) string {
	pathPrefix := c.syncInfo.dstPath
	pathSep := boscmd.BOS_PATH_SEPARATOR
	tempDstPath := ""

	if c.syncInfo.dstType == "local" {
		pathSep = util.OsPathSeparator
	} else {
		pathPrefix = c.syncInfo.dstObjectKey
	}
	// Get dst Prefix
	index := strings.LastIndex(pathPrefix, pathSep)
	if index != -1 {
		pathPrefix = pathPrefix[:index]
	}

	if pathPrefix != "" {
		tempDstPath = pathPrefix + pathSep + srcFileInfo.key
	} else {
		tempDstPath = srcFileInfo.key
	}

	if c.syncInfo.dstType == "local" {
		return replaceToOsPath(tempDstPath)
	}
	return replaceToBosPath(tempDstPath)
}
