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
	"fmt"
	// 	"os"
	// 	"runtime"
	"strings"
	"testing"
	"time"
	// 	"path/filepath"
)

import (
	"utils/util"
)

var (
	compare = &Comparator{}
)

type compareFileExistenceType struct {
	src *fileDetail
	dst *fileDetail
	ret existenceType
}

func TestCompareFileExistence(t *testing.T) {
	testCases := []compareFileExistenceType{
		compareFileExistenceType{
			src: &fileDetail{
				key: "123",
			},
			dst: &fileDetail{
				key: "123",
			},
			ret: FILE_AT_BOTH_SIDE,
		},
		compareFileExistenceType{
			src: &fileDetail{
				key: "123",
			},
			dst: &fileDetail{
				key: "223",
			},
			ret: FILE_NOT_AT_DST,
		},
		compareFileExistenceType{
			src: &fileDetail{
				key: "323",
			},
			dst: &fileDetail{
				key: "123",
			},
			ret: FILE_NOT_AT_SRC,
		},
	}
	for i, tCase := range testCases {
		ret := compare.compareFileExistence(tCase.src, tCase.dst)
		util.ExpectEqual("comparator.go compareFileExistence I", i+1, t.Errorf, tCase.ret, ret)
	}
}

type deduceDstFullPathType struct {
	syncInfo *syncArgs
	src      *fileDetail
	ret      string
}

func TestDeduceDstFullPath(t *testing.T) {
	testCases := []deduceDstFullPathType{
		// to bos
		deduceDstFullPathType{
			syncInfo: &syncArgs{
				srcPath:      "bucket1/object1/",
				dstPath:      "bos:/bucket2/object2/",
				dstObjectKey: "object2/",
				dstType:      "bos",
			},
			src: &fileDetail{
				key: "123",
			},
			ret: "object2/123",
		},
		// to bos
		deduceDstFullPathType{
			syncInfo: &syncArgs{
				srcPath:      "bucket1/object1/",
				dstPath:      "bos://bucket2/object2/",
				dstObjectKey: "object2/",
				dstType:      "bos",
			},
			src: &fileDetail{
				key: "123",
			},
			ret: "object2/123",
		},
		// to bos
		deduceDstFullPathType{
			syncInfo: &syncArgs{
				srcPath:      "bucket1/object1/",
				dstType:      "bos",
				dstObjectKey: "",
			},
			src: &fileDetail{
				key: "123",
			},
			ret: "123",
		},
		// to bos
		deduceDstFullPathType{
			syncInfo: &syncArgs{
				srcPath:      "",
				dstType:      "bos",
				dstObjectKey: "",
			},
			src: &fileDetail{
				key: "123",
			},
			ret: "123",
		},

		// to local
		deduceDstFullPathType{
			syncInfo: &syncArgs{
				srcPath: "bucket1/object1/",
				dstPath: "object2/",
				dstType: "local",
			},
			src: &fileDetail{
				key: "123",
			},
			ret: strings.Replace("object2/123", "/", util.OsPathSeparator, -1),
		},
		// to bos
		deduceDstFullPathType{
			syncInfo: &syncArgs{
				srcPath: strings.Replace("bucket1/object1/", "/", util.OsPathSeparator, -1),
				dstPath: "",
				dstType: "bos",
			},
			src: &fileDetail{
				key: "123",
			},
			ret: "123",
		},
	}
	for i, tCase := range testCases {
		compare.syncInfo = tCase.syncInfo
		ret := compare.deduceDstFullPath(tCase.src)
		util.ExpectEqual("comparator.go deduceDstFullPath I", i+1, t.Errorf, tCase.ret, ret)
	}
}

type fakeFileListIterator struct {
	fileList []listFileResult
	fileChan chan listFileResult
}

func (f *fakeFileListIterator) generateList() {
	for _, val := range f.fileList {
		f.fileChan <- val
	}
}

func (f *fakeFileListIterator) next() (*listFileResult, error) {

	select {
	case fileDetailVal := <-f.fileChan:
		return &fileDetailVal, nil
	case <-time.After(time.Duration(SYNC_COMPARATOR_TIME_OUT) * time.Millisecond):
		return nil, fmt.Errorf("Get files list time out!")
	}
	return nil, fmt.Errorf("List local files failed")
}

type comparatorType struct {
	srcIterator *fakeFileListIterator
	dstIterator *fakeFileListIterator
	atBostSize  syncStrategyInfterface
	notAtSrc    syncStrategyInfterface
	notAtDst    syncStrategyInfterface
	syncInfo    *syncArgs
	ret         []*syncOpDetail
}

func TestComparator(t *testing.T) {
	testCases := []comparatorType{
		// 1
		// at both side
		comparatorType{
			srcIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						file: &fileDetail{
							path:     "a/b/c",
							name:     "b/c",
							key:      "b/c",
							realPath: "a/b/c",
							size:     100,
							mtime:    123,
						},
					},
					listFileResult{
						file: &fileDetail{
							path:     "a/c",
							name:     "c",
							key:      "c",
							realPath: "a/c",
							size:     101,
							mtime:    124,
						},
					},
					listFileResult{
						ended: true,
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			dstIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						file: &fileDetail{
							path:     "a/b/c",
							name:     "b/c",
							key:      "b/c",
							realPath: "a/b/c",
							size:     100,
							mtime:    123,
						},
					},
					listFileResult{
						file: &fileDetail{
							path:     "a/c",
							name:     "c",
							key:      "c",
							realPath: "a/c",
							size:     101,
							mtime:    124,
						},
					},
					listFileResult{
						ended: true,
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			atBostSize: &sizeAndLastModifiedSync{},
			notAtSrc:   &deleteDstSync{},
			notAtDst:   &alwaysSync{},
			syncInfo: &syncArgs{
				srcPath:      "a/",
				dstPath:      "a/",
				srcObjectKey: "a/",
				dstObjectKey: "a/",
				srcType:      "bos",
				dstType:      "bos",
			},
			ret: []*syncOpDetail{
				&syncOpDetail{
					ended: true,
				},
			},
		},
		// 2
		// both notAtDst and botAtSrc
		comparatorType{
			srcIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						file: &fileDetail{
							path:     "a/c",
							name:     "c",
							key:      "c",
							realPath: "a/c",
							size:     100,
							mtime:    123,
						},
					},
					listFileResult{
						file: &fileDetail{
							path:     "a/d",
							name:     "d",
							key:      "d",
							realPath: "a/d",
							size:     101,
							mtime:    124,
						},
					},
					listFileResult{
						ended: true,
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			dstIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						file: &fileDetail{
							path:     "a/b",
							name:     "b",
							key:      "b",
							realPath: "a/b",
							size:     100,
							mtime:    123,
						},
					},
					listFileResult{
						file: &fileDetail{
							path:     "a/e",
							name:     "e",
							key:      "e",
							realPath: "a/e",
							size:     101,
							mtime:    124,
						},
					},
					listFileResult{
						ended: true,
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			atBostSize: &sizeAndLastModifiedSync{},
			notAtSrc:   &deleteDstSync{},
			notAtDst:   &alwaysSync{},
			syncInfo: &syncArgs{
				srcPath:      "a/",
				dstPath:      "a/",
				srcObjectKey: "a/",
				dstObjectKey: "a/",
				srcType:      "bos",
				dstType:      "bos",
			},
			ret: []*syncOpDetail{
				&syncOpDetail{
					syncFunc: OPERATE_CMD_DELETE,
					dstFileInfo: &fileDetail{
						path: "a/b",
					},
					dstPath: "a/b",
				},
				&syncOpDetail{
					syncFunc: OPERATE_CMD_COPY,
					srcFileInfo: &fileDetail{
						path: "a/c",
					},
					srcPath: "a/c",
					dstPath: "a/c",
				},
				&syncOpDetail{
					syncFunc: OPERATE_CMD_COPY,
					srcFileInfo: &fileDetail{
						path: "a/d",
					},
					srcPath: "a/d",
					dstPath: "a/d",
				},
				&syncOpDetail{
					syncFunc: OPERATE_CMD_DELETE,
					dstFileInfo: &fileDetail{
						path: "a/e",
					},
					dstPath: "a/e",
				},
				&syncOpDetail{
					ended: true,
				},
			},
		},
		// 3
		// at both side and not at dst
		comparatorType{
			srcIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						file: &fileDetail{
							path:     "a/b/c",
							name:     "b/c",
							key:      "b/c",
							realPath: "a/b/c",
							size:     100,
							mtime:    123,
						},
					},
					listFileResult{
						file: &fileDetail{
							path:     "a/c",
							name:     "c",
							key:      "c",
							realPath: "a/c",
							size:     101,
							mtime:    124,
						},
					},
					listFileResult{
						ended: true,
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			dstIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						file: &fileDetail{
							path:     "a/b/c",
							name:     "b/c",
							key:      "b/c",
							realPath: "a/b/c",
							size:     100,
							mtime:    121,
						},
					},
					listFileResult{
						ended: true,
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			atBostSize: &sizeAndLastModifiedSync{},
			notAtSrc:   &deleteDstSync{},
			notAtDst:   &alwaysSync{},
			syncInfo: &syncArgs{
				srcPath:      "a/",
				dstPath:      "a/",
				srcObjectKey: "a/",
				dstObjectKey: "a/",
				srcType:      "bos",
				dstType:      "bos",
			},
			ret: []*syncOpDetail{
				&syncOpDetail{
					syncFunc: OPERATE_CMD_COPY,
					srcFileInfo: &fileDetail{
						path: "a/b/c",
					},
					dstFileInfo: &fileDetail{
						path: "a/b/c",
					},
					srcPath: "a/b/c",
					dstPath: "a/b/c",
				},
				&syncOpDetail{
					syncFunc: OPERATE_CMD_COPY,
					srcFileInfo: &fileDetail{
						path: "a/c",
					},
					srcPath: "a/c",
					dstPath: "a/c",
				},
				&syncOpDetail{
					ended: true,
				},
			},
		},
		// 4
		// not at src
		comparatorType{
			srcIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						file: &fileDetail{
							path:     "a/b/c",
							name:     "b/c",
							key:      "b/c",
							realPath: "a/b/c",
							size:     100,
							mtime:    121,
						},
					},
					listFileResult{
						ended: true,
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			dstIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						file: &fileDetail{
							path:     "a/b/c",
							name:     "b/c",
							key:      "b/c",
							realPath: "a/b/c",
							size:     100,
							mtime:    123,
						},
					},
					listFileResult{
						file: &fileDetail{
							path:     "a/c",
							name:     "c",
							key:      "c",
							realPath: "a/c",
							size:     101,
							mtime:    124,
						},
					},
					listFileResult{
						ended: true,
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			atBostSize: &sizeAndLastModifiedSync{},
			notAtSrc:   &deleteDstSync{},
			notAtDst:   &alwaysSync{},
			syncInfo: &syncArgs{
				srcPath:      "a/",
				dstPath:      "a/",
				srcObjectKey: "a/",
				dstObjectKey: "a/",
				srcType:      "bos",
				dstType:      "bos",
			},
			ret: []*syncOpDetail{
				&syncOpDetail{
					syncFunc: OPERATE_CMD_DELETE,
					dstFileInfo: &fileDetail{
						path: "a/c",
					},
					dstPath: "a/c",
				},
				&syncOpDetail{
					ended: true,
				},
			},
		},
		// 5
		// error occurs
		comparatorType{
			srcIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						file: &fileDetail{
							path:     "a/b/c",
							name:     "b/c",
							key:      "b/c",
							realPath: "a/b/c",
							size:     100,
							mtime:    121,
						},
					},
					listFileResult{
						err: fmt.Errorf("error 1"),
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			dstIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						file: &fileDetail{
							path:     "a/b/c",
							name:     "b/c",
							key:      "b/c",
							realPath: "a/b/c",
							size:     100,
							mtime:    123,
						},
					},
					listFileResult{
						file: &fileDetail{
							path:     "a/c",
							name:     "c",
							key:      "c",
							realPath: "a/c",
							size:     101,
							mtime:    124,
						},
					},
					listFileResult{
						ended: true,
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			atBostSize: &sizeAndLastModifiedSync{},
			notAtSrc:   &deleteDstSync{},
			notAtDst:   &alwaysSync{},
			syncInfo: &syncArgs{
				srcPath:      "a/",
				dstPath:      "a/",
				srcObjectKey: "a/",
				dstObjectKey: "a/",
				srcType:      "bos",
				dstType:      "bos",
			},
			ret: []*syncOpDetail{
				&syncOpDetail{
					err: fmt.Errorf("error 1"),
				},
			},
		},
		// 6
		// error occurs
		comparatorType{
			srcIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						err: fmt.Errorf("error 1"),
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			dstIterator: &fakeFileListIterator{
				fileList: []listFileResult{
					listFileResult{
						ended: true,
					},
				},
				fileChan: make(chan listFileResult, 1),
			},
			atBostSize: &sizeAndLastModifiedSync{},
			notAtSrc:   &deleteDstSync{},
			notAtDst:   &alwaysSync{},
			syncInfo: &syncArgs{
				srcPath:      "a/",
				dstPath:      "a/",
				srcObjectKey: "a/",
				dstObjectKey: "a/",
				srcType:      "bos",
				dstType:      "bos",
			},
			ret: []*syncOpDetail{
				&syncOpDetail{
					err: fmt.Errorf("error 1"),
				},
			},
		},
	}
	for i, tCase := range testCases {
		go tCase.srcIterator.generateList()
		go tCase.dstIterator.generateList()
		com := &Comparator{
			fileAtBothSideSyncStrategy: tCase.atBostSize,
			fileNotAtSrcSyncStrategy:   tCase.notAtSrc,
			fileNotAtDstSyncStrategy:   tCase.notAtDst,
			syncOpChan:                 make(chan syncOpDetail, 1),
			syncInfo:                   tCase.syncInfo,
		}
		go com.compare(tCase.srcIterator, tCase.dstIterator)
		index := 0
		for {
			syncOp, err := com.next()
			if err != nil {
				t.Logf("I id %d err %s", i+1, err)
				util.ExpectEqual("comparator.go comparator I", i+1, t.Errorf,
					tCase.ret[index] == nil, true)
				break
			}
			if syncOp.err != nil {
				t.Logf("II id %d err %s", i+1, syncOp.err)
				util.ExpectEqual("comparator.go comparator II", i+1, t.Errorf,
					tCase.ret[index].err != nil, true)
				break
			}
			if syncOp.ended {
				util.ExpectEqual("comparator.go comparator III", i+1, t.Errorf,
					tCase.ret[index].ended, true)
				break
			}
			util.ExpectEqual("comparator.go comparator IV", i+1, t.Errorf,
				tCase.ret[index].syncFunc, syncOp.syncFunc)
			if syncOp.syncFunc != OPERATE_CMD_DELETE {
				util.ExpectEqual("comparator.go comparator V", i+1, t.Errorf,
					tCase.ret[index].srcPath, syncOp.srcPath)
			}

			util.ExpectEqual("comparator.go comparator VI", i+1, t.Errorf,
				tCase.ret[index].dstPath, syncOp.dstPath)
			index++
		}
	}
}
