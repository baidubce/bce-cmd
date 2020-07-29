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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

import (
	"bceconf"
	"github.com/baidubce/bce-sdk-go/util/log"
	"utils/util"
)

const (
	SRC_IS_BOS    = "srcIsBos"
	SRC_IS_LOCAL  = "srcIsLocal"
	SRC_IS_STREAM = "srcIsStream"
	MD5_CAlC_SIZE = 1024 * 1024
	FLUSH_PERIOD  = 20 * time.Second
)

type CompletePartInfo struct {
	PartNumberId int64  `json:"partNumberId"`
	ETag         string `json:"eTag"`
}

type BreakPointRecord struct {
	Md5Val              string             `json:"md5"`
	UploadId            string             `json:"uploadId"`
	SrcFileSize         int64              `json:"fileSize"`
	SrcFileLastModified int64              `json:"fileModifyTime"`
	PartsNum            int64              `json:"partsNum"`
	PartSize            int64              `json:"partSize"`
	CompltePartList     []CompletePartInfo `json:"completePartList"`
	RecordTime          int64              `json:"recordTime"`
}

type MultiTaskContent struct {
	contentId           string
	srcType             string // bos or local file
	dstType             string
	srcFilePath         string
	dstFilePath         string
	srcBucketName       string
	srcObjectKey        string
	dstBucketName       string
	dstObjectKey        string
	md5Val              string
	uploadId            string
	breakPointPath      string
	srcFileSize         int64
	srcFileLastModified int64
	partSize            int64
	partsNum            int64
	initPartSize        int64
	lastFlushTime       int64
	lastFinshPartTime   int64
	compltePartNum      int
	compltePartList     map[int64]CompletePartInfo
	isDirty             bool
	needRestart         bool
	rwmutex             sync.RWMutex
	flushLock           sync.Mutex
}

func (m *MultiTaskContent) init(srcBucketName, srcObjectKey, dstBucketName, dstObjectKey,
	srcType, dstType, md5Val string, fileSize, fileMtime int64, restart bool,
	initPartSize int64) error {

	m.srcType = srcType
	m.dstType = dstType
	m.srcBucketName = srcBucketName
	m.srcObjectKey = srcObjectKey
	m.dstBucketName = dstBucketName
	m.dstObjectKey = dstObjectKey
	m.isDirty = false
	m.needRestart = true
	m.compltePartList = make(map[int64]CompletePartInfo)
	m.initPartSize = initPartSize

	if srcType == IS_LOCAL {
		m.srcFilePath = srcObjectKey
	} else if srcType == IS_BOS {
		m.srcFilePath = BOS_PATH_PREFIX + srcBucketName + util.BOS_PATH_SEPARATOR + srcObjectKey
	}

	if dstType == IS_LOCAL {
		m.dstFilePath = dstObjectKey
	} else if dstType == IS_BOS {
		m.dstFilePath = BOS_PATH_PREFIX + dstBucketName + util.BOS_PATH_SEPARATOR + dstObjectKey
	}

	m.srcFileSize = fileSize
	m.srcFileLastModified = fileMtime
	m.md5Val = md5Val

	// calc the file path of breakPointRecord
	m.contentId = util.StringMd5(m.srcFilePath + "_" + m.dstFilePath)
	m.breakPointPath = filepath.Join(bceconf.MultiuploadFolder, m.contentId)

	// mkdir MultiuploadFolder
	if !util.DoesDirExist(bceconf.MultiuploadFolder) {
		if err := util.TryMkdir(bceconf.MultiuploadFolder); err != nil {
			return err
		}
	}

	// if restart, we need to delete the old breakPointRecord file.
	// otherwise, reading info from breakPointRecord
	m.partSize = -1
	if !restart && util.DoesFileExist(m.breakPointPath) {
		breakPointTemp := &BreakPointRecord{}

		fd, err := os.Open(m.breakPointPath)
		if err != nil {
			return err
		}

		breakPointDecoder := json.NewDecoder(fd)
		if err = breakPointDecoder.Decode(breakPointTemp); err != nil {
			return fmt.Errorf("can not get breakPointRecord from path %s", m.breakPointPath)
		}

		m.uploadId = breakPointTemp.UploadId
		if m.breakPointRecordIsValid(breakPointTemp) {
			log.Debugf("%s => %s, have breakpoint record and it is vaild", m.srcFilePath,
				m.dstFilePath)
			for _, val := range breakPointTemp.CompltePartList {
				if val.ETag == "" {
					log.Debugf("part %d have empty etag %s, when %s => %s", val.PartNumberId,
						m.srcFilePath, m.dstFilePath)
					continue
				}
				m.compltePartList[val.PartNumberId] = CompletePartInfo{
					PartNumberId: val.PartNumberId,
					ETag:         val.ETag,
				}
				m.compltePartNum++
			}
			m.partSize = breakPointTemp.PartSize
			m.partsNum = breakPointTemp.PartsNum
			m.lastFlushTime = breakPointTemp.RecordTime
			m.lastFinshPartTime = breakPointTemp.RecordTime
			m.needRestart = false
		} else {
			log.Debugf("%s => %s, have breakpoint record and it is not vaild", m.srcFilePath,
				m.dstFilePath)
		}
	}

	if m.needRestart {
		m.Remove()
		m.clear()
		m.getPartInfo()
	}

	// start background flush
	go m.startBackgroundFlush()

	return nil
}

func (m *MultiTaskContent) clear() {
	m.uploadId = ""
	m.partSize = 0
	m.partsNum = 0
	m.isDirty = false
	m.lastFlushTime = 0
	m.compltePartNum = 0
	m.lastFinshPartTime = 0
	m.compltePartList = make(map[int64]CompletePartInfo)
}

func (m *MultiTaskContent) breakPointRecordIsValid(record *BreakPointRecord) bool {
	if m.srcFileSize != record.SrcFileSize {
		return false
	}
	if m.srcFileLastModified != record.SrcFileLastModified {
		return false
	}
	if m.md5Val != record.Md5Val || record.UploadId == "" {
		return false
	}
	// is download
	if m.dstType == IS_LOCAL {
		if !util.DoesFileExist(record.UploadId) {
			return false
		}
	}

	breakPointExpireTime, _ := bceconf.ServerConfigProvider.GetBreakpointFileExpiration()
	breakPointExpireTime *= 86400
	if time.Now().Unix()-record.RecordTime > int64(breakPointExpireTime) {
		return false
	}
	return true
}

func (m *MultiTaskContent) finishPart(partNumber int64, eTag string) error {
	if partNumber < 1 || partNumber > m.partsNum {
		return fmt.Errorf("part number %d is invalid! part number not in [1, %d]", m.partsNum)
	} else if eTag == "" {
		return fmt.Errorf("the etag of part %d is empty", partNumber)
	}

	m.rwmutex.Lock()
	defer m.rwmutex.Unlock()
	if val, ok := m.compltePartList[partNumber]; ok {
		log.Debugf("%s => %s, part number %d upload again, old part Inof is%v", m.srcFilePath,
			m.dstFilePath, partNumber, val)
	} else {
		m.compltePartNum++
	}
	m.compltePartList[partNumber] = CompletePartInfo{
		PartNumberId: partNumber,
		ETag:         eTag,
	}
	m.isDirty = true
	m.lastFinshPartTime = time.Now().Unix()
	return nil
}

func (m *MultiTaskContent) initSrcFileInfoFromLocal(filePath string) error {
	// Get the info of src file agin
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	m.srcFileSize = info.Size()
	m.srcFileLastModified = info.ModTime().Unix()

	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.Seek(MD5_CAlC_SIZE, 2)
	if err != nil {
		return err
	}

	md5New := md5.New()
	copied, err := io.CopyN(md5New, fd, MD5_CAlC_SIZE)

	if err != nil {
		return err
	} else if copied != MD5_CAlC_SIZE {
		return fmt.Errorf("Failed to calc md5 of src path %s", filePath)
	}
	m.md5Val = hex.EncodeToString(md5New.Sum(nil))
	return nil
}

func (m *MultiTaskContent) getPartInfo() {
	partSize := m.initPartSize

	if partSize*MAX_PARTS < m.srcFileSize {
		lowerLimit := int64(math.Ceil(float64(m.srcFileSize) / MAX_PARTS))
		partSize = int64(math.Ceil(float64(lowerLimit)/float64(partSize))) * partSize
	}

	partsNum := (m.srcFileSize + partSize - 1) / partSize
	m.partSize = partSize
	m.partsNum = partsNum
}

func (m *MultiTaskContent) partIsFinish(partNumber int64) (*CompletePartInfo, bool) {
	m.rwmutex.RLock()
	defer m.rwmutex.RUnlock()
	if val, ok := m.compltePartList[partNumber]; ok {
		if m.compltePartList[partNumber].ETag == "" {
			delete(m.compltePartList, partNumber)
			m.compltePartNum--
		} else {
			return &val, true
		}
	}
	return nil, false
}

func (m *MultiTaskContent) complete() error {
	m.isDirty = false
	err := m.Remove()
	if err == nil || ErrIsNotExist(err) {
		return nil
	}
	return err
}

// make a snapshot for breakpoint transmission
func (m *MultiTaskContent) snapshot() *BreakPointRecord {
	m.rwmutex.RLock()
	defer m.rwmutex.RUnlock()

	breakPointTemp := &BreakPointRecord{
		Md5Val:              m.md5Val,
		UploadId:            m.uploadId,
		SrcFileSize:         m.srcFileSize,
		SrcFileLastModified: m.srcFileLastModified,
		PartsNum:            m.partsNum,
		PartSize:            m.partSize,
		RecordTime:          m.lastFinshPartTime,
		CompltePartList:     make([]CompletePartInfo, len(m.compltePartList)),
	}

	i := 0
	for k := range m.compltePartList {
		breakPointTemp.CompltePartList[i] = m.compltePartList[k]
		i++
	}
	return breakPointTemp
}

// flush the record of breakpoint transmission to disk
func (m *MultiTaskContent) flush(breakPointTemp *BreakPointRecord) error {
	m.flushLock.Lock()
	defer m.flushLock.Unlock()

	fd, err := os.OpenFile(m.breakPointPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer fd.Close()

	bytes, err := json.Marshal(breakPointTemp)
	if err != nil {
		return err
	}
	if _, err = fd.Write(bytes); err != nil {
		return err
	}
	m.lastFlushTime = breakPointTemp.RecordTime
	return nil
}

// flush the record of breakpoint transmission to disk every FLUSH_PERIOD seconds
func (m *MultiTaskContent) startBackgroundFlush() {
	for {
		time.Sleep(FLUSH_PERIOD)
		if !m.isDirty || m.lastFlushTime == m.lastFinshPartTime {
			continue
		}
		breakPointTemp := m.snapshot()
		m.flush(breakPointTemp)
	}
}

// flush the record of breakpoint transmission to disk
func (m *MultiTaskContent) Flush() error {
	if !m.isDirty || m.lastFlushTime == m.lastFinshPartTime {
		return nil
	}
	breakPointTemp := m.snapshot()
	return m.flush(breakPointTemp)
}

// delete
func (m *MultiTaskContent) Remove() error {
	m.flushLock.Lock()
	defer m.flushLock.Unlock()

	m.isDirty = false

	if util.DoesFileExist(m.breakPointPath) {
		return os.Remove(m.breakPointPath)
	}

	if m.dstType == IS_LOCAL && util.DoesFileExist(m.uploadId) {
		return os.Remove(m.uploadId)
	}
	return nil
}

// flush the record of breakpoint transmission to disk when bcecmd non-mormal exit
func (m *MultiTaskContent) Exit() error {
	if !m.isDirty || m.lastFlushTime == m.lastFinshPartTime {
		return nil
	}
	breakPointTemp := m.snapshot()
	return m.flush(breakPointTemp)
}

func (m *MultiTaskContent) GetId() (string, error) {
	if m.contentId == "" {
		return "", fmt.Errorf("content ID is empty!")
	}
	return m.contentId, nil
}

func (m *MultiTaskContent) GetFinshPartNum() int {
	return m.compltePartNum
}
