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
	"path/filepath"
	"testing"
)

import (
	"utils/util"
)

type sizeAndLastModifiedSyncType struct {
	src *fileDetail
	dst *fileDetail
	ret bool
}

func TestSizeAndLastModifiedSync(t *testing.T) {
	testCases := []sizeAndLastModifiedSyncType{
		sizeAndLastModifiedSyncType{
			src: &fileDetail{
				mtime: 123,
				size:  100,
			},
			dst: &fileDetail{
				mtime: 123,
				size:  100,
			},
			ret: false,
		},
		sizeAndLastModifiedSyncType{
			src: &fileDetail{
				mtime: 122,
				size:  100,
			},
			dst: &fileDetail{
				mtime: 123,
				size:  100,
			},
			ret: false,
		},
		sizeAndLastModifiedSyncType{
			src: &fileDetail{
				mtime: 124,
				size:  100,
			},
			dst: &fileDetail{
				mtime: 123,
				size:  100,
			},
			ret: true,
		},
		sizeAndLastModifiedSyncType{
			src: &fileDetail{
				mtime: 124,
				size:  100,
			},
			dst: &fileDetail{
				mtime: 124,
				size:  101,
			},
			ret: true,
		},
	}
	strategyBoth := &sizeAndLastModifiedSync{}
	strategyDel := &deleteDstSync{}
	strategyNever := &neverSync{}
	strategyAlway := &alwaysSync{}
	for i, tCase := range testCases {
		ret, _ := strategyBoth.shouldSync(tCase.src, tCase.dst)
		util.ExpectEqual("tools.go sizeAndLastModifiedSync I", i+1, t.Errorf, tCase.ret, ret)
		ret, _ = strategyDel.shouldSync(nil, tCase.dst)
		util.ExpectEqual("tools.go deleteDstSync I", i+1, t.Errorf, true, ret)
		ret, _ = strategyDel.shouldSync(tCase.src, tCase.dst)
		util.ExpectEqual("tools.go deleteDstSync II", i+1, t.Errorf, false, ret)
		ret, _ = strategyDel.shouldSync(tCase.src, nil)
		util.ExpectEqual("tools.go deleteDstSync III", i+1, t.Errorf, false, ret)
		ret, _ = strategyNever.shouldSync(tCase.src, tCase.dst)
		ret, _ = strategyDel.shouldSync(nil, nil)
		util.ExpectEqual("tools.go deleteDstSync IV", i+1, t.Errorf, false, ret)
		util.ExpectEqual("tools.go neverSync I", i+1, t.Errorf, false, ret)
		ret, _ = strategyNever.shouldSync(tCase.src, nil)
		util.ExpectEqual("tools.go neverSync II", i+1, t.Errorf, false, ret)
		ret, _ = strategyNever.shouldSync(nil, tCase.dst)
		util.ExpectEqual("tools.go neverSync III", i+1, t.Errorf, false, ret)
		ret, _ = strategyNever.shouldSync(nil, nil)
		util.ExpectEqual("tools.go neverSync IV", i+1, t.Errorf, false, ret)
		ret, _ = strategyAlway.shouldSync(tCase.src, tCase.dst)
		util.ExpectEqual("tools.go strategyAlway I", i+1, t.Errorf, true, ret)
		ret, _ = strategyAlway.shouldSync(tCase.src, nil)
		util.ExpectEqual("tools.go strategyAlway II", i+1, t.Errorf, true, ret)
		ret, _ = strategyAlway.shouldSync(nil, tCase.dst)
		util.ExpectEqual("tools.go strategyAlway III", i+1, t.Errorf, false, ret)
		ret, _ = strategyAlway.shouldSync(nil, nil)
		util.ExpectEqual("tools.go strategyAlway IV", i+1, t.Errorf, false, ret)
	}
	util.ExpectEqual("tools.go sizeAndLastModifiedSync gen", 1, t.Errorf, OPERATE_CMD_COPY,
		strategyBoth.genSyncFunc())
	util.ExpectEqual("tools.go deleteDstSync gen", 1, t.Errorf, OPERATE_CMD_DELETE,
		strategyDel.genSyncFunc())
	util.ExpectEqual("tools.go neverSync gen", 1, t.Errorf, OPERATE_CMD_NOTHING,
		strategyNever.genSyncFunc())
	util.ExpectEqual("tools.go alwaysSync gen", 1, t.Errorf, OPERATE_CMD_COPY,
		strategyAlway.genSyncFunc())
}

type deleteDstSyncType struct {
	dstType       string
	dstBucketName string
	dst           *fileDetail
	src           *fileDetail
	excludeDelete []string
	ret           bool
}

func TestDeleteDstSyncType(t *testing.T) {
	testCases := []deleteDstSyncType{
		deleteDstSyncType{
			src: &fileDetail{
				path: "./.git/http",
			},
			dst: &fileDetail{
				path: "./.git/http",
			},
			ret: false,
		},
		deleteDstSyncType{
			src: nil,
			dst: nil,
			ret: false,
		},
		deleteDstSyncType{
			src: nil,
			dst: &fileDetail{
				path: "./.git/http",
			},
			ret: true,
		},
		//4
		deleteDstSyncType{
			dstType:       IS_LOCAL,
			excludeDelete: []string{"./.git"},
			src:           nil,
			dst: &fileDetail{
				path: "./.git/http",
			},
			ret: true,
		},
		deleteDstSyncType{
			dstType:       IS_LOCAL,
			excludeDelete: []string{"./.git"},
			src:           nil,
			dst: &fileDetail{
				path: "./.git",
			},
			ret: false,
		},
		deleteDstSyncType{
			dstType:       IS_LOCAL,
			excludeDelete: []string{"./.git/*"},
			src:           nil,
			dst: &fileDetail{
				path: "./.git/http",
			},
			ret: false,
		},
		//7
		deleteDstSyncType{
			dstType:       IS_LOCAL,
			excludeDelete: []string{"/home/work/.git/*"},
			src:           nil,
			dst: &fileDetail{
				path: "/home/work/.git/http",
			},
			ret: false,
		},
		deleteDstSyncType{
			dstType:       IS_BOS,
			dstBucketName: "test",
			excludeDelete: []string{"bos:/test/.git"},
			src:           nil,
			dst: &fileDetail{
				path: ".git/http",
			},
			ret: true,
		},
		deleteDstSyncType{
			dstType:       IS_BOS,
			dstBucketName: "test",
			excludeDelete: []string{"bos:/test/.git"},
			src:           nil,
			dst: &fileDetail{
				path: ".git/http",
			},
			ret: true,
		},
		deleteDstSyncType{
			dstType:       IS_BOS,
			dstBucketName: "test",
			excludeDelete: []string{"bos:/test/.git"},
			src:           nil,
			dst: &fileDetail{
				path: ".git",
			},
			ret: false,
		},
		deleteDstSyncType{
			dstType:       IS_BOS,
			dstBucketName: "test",
			excludeDelete: []string{"bos:/test/.git/*"},
			src:           nil,
			dst: &fileDetail{
				path: ".git",
			},
			ret: true,
		},
		deleteDstSyncType{
			dstType:       IS_BOS,
			dstBucketName: "test",
			excludeDelete: []string{"bos:/test/.git/*"},
			src:           nil,
			dst: &fileDetail{
				path: ".git/http",
			},
			ret: false,
		},
	}
	for i, tCase := range testCases {
		var (
			deleteFilter *bosFilter = nil
			retCode      BosCliErrorCode
			err          error
		)

		if len(tCase.excludeDelete) > 0 {
			deleteFilter, retCode, err = newSyncFilter(tCase.excludeDelete, []string{}, []string{},
				[]string{}, tCase.dstType == IS_LOCAL)
			util.ExpectEqual("tools.go deleteDstSync I", i+1, t.Errorf, retCode, BOSCLI_OK)
			util.ExpectEqual("tools.go deleteDstSync II", i+1, t.Errorf, err, nil)
		}

		strategyDel := &deleteDstSync{
			deleteFilter:  deleteFilter,
			dstType:       tCase.dstType,
			dstBucketName: tCase.dstBucketName,
		}

		if tCase.dstType == IS_LOCAL {
			tCase.dst.path, _ = filepath.Abs(tCase.dst.path)
		}

		ret, _ := strategyDel.shouldSync(tCase.src, tCase.dst)
		util.ExpectEqual("tools.go deleteDstSync II", i+1, t.Errorf, tCase.ret, ret)
	}
}
