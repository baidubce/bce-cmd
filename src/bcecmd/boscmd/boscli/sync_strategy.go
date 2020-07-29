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

// This module provides the major sync strategies for BOS.

package boscli

import (
	"bcecmd/boscmd"
	"github.com/baidubce/bce-sdk-go/util/log"
)

type syncStrategyInfterface interface {
	shouldSync(*fileDetail, *fileDetail) (bool, error)
	genSyncFunc() string
}

// only compare crc32
type crc32Sync struct {
	srcType       string
	dstType       string
	srcBucketName string
	dstBucketName string
	srcBosClient  bosClientInterface
	dstBosClient  bosClientInterface
}

func (s *crc32Sync) shouldSync(src *fileDetail, dst *fileDetail) (bool, error) {
	var (
		srcCrc32Val string
		dstCrc32Val string
		err         error
	)

	if s.srcType == IS_LOCAL {
		if srcCrc32Val, err = getCrc32OfLocalFile(src.path); err != nil {
			return false, err
		}
	} else {
		if srcObjectMeta, err := getObjectMeta(s.srcBosClient, s.srcBucketName,
			src.path); err != nil {
			return false, err
		} else {
			srcCrc32Val = srcObjectMeta.crc32
		}

	}

	if s.dstType == IS_LOCAL {
		if dstCrc32Val, err = getCrc32OfLocalFile(dst.path); err != nil {
			return false, err
		}
	} else {
		if dstObjectMeta, err := getObjectMeta(s.dstBosClient, s.dstBucketName,
			dst.path); err != nil {
			return false, err
		} else {
			dstCrc32Val = dstObjectMeta.crc32
		}

	}

	if srcCrc32Val != "" && srcCrc32Val == dstCrc32Val {
		log.Debugf("src path: %s, dst path %s src crc32 is : %s, dst crc32 is %s should not sync",
			src.path, dst.path, srcCrc32Val, dstCrc32Val)
		return false, nil
	} else {
		// this object don't have crc32 in bos
		log.Debugf("src path: %s, dst path %s src crc32 is : %s, dst crc32 is %s should sync",
			src.path, dst.path, srcCrc32Val, dstCrc32Val)
		return true, nil
	}
}

func (s *crc32Sync) genSyncFunc() string {
	return OPERATE_CMD_COPY
}

// compare size and last modified at first, then compare crc32
type sizeAndLastModifiedAndCrc32Sync struct {
	srcType       string
	dstType       string
	srcBucketName string
	dstBucketName string
	srcBosClient  bosClientInterface
	dstBosClient  bosClientInterface
}

func (s *sizeAndLastModifiedAndCrc32Sync) shouldSync(src *fileDetail, dst *fileDetail) (bool, error) {

	srcMtime := src.mtime
	dstMtime := dst.mtime
	srcSize := src.size
	dstSize := dst.size

	if srcMtime < dstMtime || (srcMtime == dstMtime && srcSize == dstSize) {
		log.Debugf("src path: %s, dst path %s src size: %d, dst size %d src lastModified: %d, dst "+
			"lastModified %d should not sync", src.path, dst.path, srcSize, dstSize, srcMtime, dstMtime)
		return false, nil
	}

	crc32SyncStrategy := crc32Sync{
		srcType:       s.srcType,
		dstType:       s.dstType,
		srcBucketName: s.srcBucketName,
		dstBucketName: s.dstBucketName,
		srcBosClient:  s.srcBosClient,
		dstBosClient:  s.dstBosClient,
	}
	return crc32SyncStrategy.shouldSync(src, dst)
}

func (s *sizeAndLastModifiedAndCrc32Sync) genSyncFunc() string {
	return OPERATE_CMD_COPY
}

type sizeAndLastModifiedSync struct{}

// Compares size and last modified time only
func (s *sizeAndLastModifiedSync) shouldSync(src *fileDetail, dst *fileDetail) (bool, error) {

	srcMtime := src.mtime
	dstMtime := dst.mtime
	srcSize := src.size
	dstSize := dst.size

	if srcMtime > dstMtime || (srcMtime == dstMtime && srcSize != dstSize) {

		log.Debugf("src path: %s, dst path %s src size: %d, dst size %d src lastModified: %d, dst "+
			"lastModified %d should sync", src.key, dst.key, srcSize, dstSize, srcMtime, dstMtime)
		return true, nil
	}
	log.Debugf("src path: %s, dst path %s src size: %d, dst size %d src lastModified: %d, dst "+
		"lastModified %d should not sync", src.key, dst.key, srcSize, dstSize, srcMtime, dstMtime)
	return false, nil
}

func (s *sizeAndLastModifiedSync) genSyncFunc() string {
	return OPERATE_CMD_COPY
}

// Deletes the sync dst
type deleteDstSync struct {
	deleteFilter  *bosFilter
	dstType       string
	dstBucketName string
}

func (d *deleteDstSync) shouldSync(src *fileDetail, dst *fileDetail) (bool, error) {
	if src != nil {
		log.Debugf("src object %s existing, should not delete dst object", src.path)
		return false, nil
	} else if dst == nil {
		log.Debugf("dst is nil, should not delete dst object")
		return false, nil
	} else if d.deleteFilter == nil {
		log.Debugf("delete dst %s", dst.path)
		return true, nil
	}

	// pattern filter
	dstPath := dst.path
	if d.dstType == IS_BOS {
		dstPath = d.dstBucketName + boscmd.BOS_PATH_SEPARATOR + dst.path
	}

	filtered, err := d.deleteFilter.PatternFilter(dstPath)
	if err != nil {
		return false, err
	} else if filtered {
		log.Debugf("dst %s is filtered out, shuld not delete ", dst.path)
		return false, nil
	}
	log.Debugf("dst %s not be filtered out, shuld delete ", dst.path)
	return true, nil
}

func (s *deleteDstSync) genSyncFunc() string {
	return OPERATE_CMD_DELETE
}

// Does nothing
type neverSync struct{}

func (n *neverSync) shouldSync(src *fileDetail, dst *fileDetail) (bool, error) {
	log.Debugf("should not sync")
	return false, nil
}

func (n *neverSync) genSyncFunc() string {
	return OPERATE_CMD_NOTHING
}

// Always sync the src to the dst
type alwaysSync struct{}

func (n *alwaysSync) shouldSync(src *fileDetail, dst *fileDetail) (bool, error) {
	if src != nil {
		log.Debugf("should sync")
		return true, nil
	}
	log.Debugf("should not sync")
	return false, nil
}

func (n *alwaysSync) genSyncFunc() string {
	return OPERATE_CMD_COPY
}
