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

// This file is a bridge between boscli and gosdk.

package boscli

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

import (
	"bcecmd/boscmd"
	"bceconf"
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
	"github.com/baidubce/bce-sdk-go/util/log"
	"utils/util"
)

const (
	MAX_DELETE_NUM_EACH_TIME = 500
)

// process upload, download, copy or delete
type cliHandler struct{}

// List object
type objectListIterator struct {
	filter      *bosFilter
	bosClient   bosClientInterface
	objectsChan chan listFileResult
}

// all: show all objects
// recursive: don't show pre
// full: show full path of object (don't contain bucket name)
// isObject: represent objectKey point to a single object, just get info of it
func NewObjectListIterator(bosClient bosClientInterface, filter *bosFilter, bucketName,
	objectKey, marker string, all, recursive, srcIsDir, showEmptyDir bool,
	maxKeys int) *objectListIterator {

	objects := &objectListIterator{
		filter:      filter,
		bosClient:   bosClient,
		objectsChan: make(chan listFileResult, maxKeys),
	}
	if srcIsDir {
		go objects.listAllObjects(bucketName, objectKey, marker, all, recursive, maxKeys,
			showEmptyDir)
	} else {
		go objects.getSingleObjectInfo(bucketName, objectKey, showEmptyDir)
	}
	return objects
}

func (o *objectListIterator) getSingleObjectInfo(bucketName, objectKey string, showEmptyDir bool) {

	trimPos := strings.LastIndex(objectKey, boscmd.BOS_PATH_SEPARATOR) + 1

	flieInfo, err := getObjectMeta(o.bosClient, bucketName, objectKey)

	if err != nil {
		o.objectsChan <- listFileResult{err: err}
		return
	}
	// this object may be an empty dir
	if !showEmptyDir || len(objectKey) >= trimPos {
		flieInfo.key = objectKey[trimPos:]
	} else {
		flieInfo.key = objectKey
	}

	o.objectsChan <- listFileResult{
		file: flieInfo,
	}
	o.objectsChan <- listFileResult{
		ended: true,
	}
}

func (o *objectListIterator) listAllObjects(bucketName, objectKey, marker string, all,
	recursive bool, maxKeys int, showEmptyDir bool) {

	var (
		delimiter           string = "/"
		trimPos             int    = 0
		objectKeyIsEmptyDir bool
		objectKeyDetail     *listFileResult
	)

	if recursive {
		delimiter = ""
	}

	trimPos = strings.LastIndex(objectKey, boscmd.BOS_PATH_SEPARATOR) + 1

	for {
		args := &api.ListObjectsArgs{
			Delimiter: delimiter,
			Marker:    marker,
			MaxKeys:   maxKeys,
			Prefix:    objectKey,
		}
		response, err := o.bosClient.ListObjects(bucketName, args)
		gtime := time.Now().Unix()

		if err != nil {
			o.objectsChan <- listFileResult{err: err}
			goto END
		}

		// throw dir
		if len(response.CommonPrefixes) != 0 {
			for _, item := range response.CommonPrefixes {
				o.objectsChan <- listFileResult{
					dir: &dirDetail{
						path: item.Prefix,
						key:  item.Prefix[trimPos:],
					}, isDir: true,
				}
			}
		}

		// throw objects
		for _, item := range response.Contents {
			// this object may be an empty dir
			if strings.HasSuffix(item.Key, boscmd.BOS_PATH_SEPARATOR) && !recursive {
				continue
			}

			// utc to local timestamp
			modiTime, err := util.TranUTCTimeStringToTimeStamp(item.LastModified, BOS_TIME_FORMT)
			if err != nil {
				o.objectsChan <- listFileResult{err: err}
				goto END
			}

			// objectKey may be an empty dir
			if recursive && strings.TrimSpace(item.Key[trimPos:]) == "" {
				// we should throw this empty dir when
				if showEmptyDir {
					objectKeyIsEmptyDir = true
					objectKeyDetail = &listFileResult{
						file: &fileDetail{
							path:         item.Key,
							key:          "",
							mtime:        modiTime,
							gtime:        gtime,
							size:         int64(item.Size),
							storageClass: item.StorageClass,
						},
						isDir: strings.HasSuffix(item.Key, boscmd.BOS_PATH_SEPARATOR),
					}
				}
				continue
			}

			// pattern filter
			if o.filter != nil {
				bosPath := bucketName + boscmd.BOS_PATH_SEPARATOR + item.Key
				if filtered, err := o.filter.PatternFilter(bosPath); filtered || err != nil {
					// only error is ErrBadPattern, when it occurs we need to stop list local file
					if err != nil {
						o.objectsChan <- listFileResult{err: err}
						goto END
					}
					continue
				}
			}

			// time filter
			if o.filter != nil && o.filter.TimeFilter(modiTime) {
				continue
			}

			o.objectsChan <- listFileResult{
				file: &fileDetail{
					path:         item.Key,
					key:          item.Key[trimPos:],
					mtime:        modiTime,
					gtime:        gtime,
					size:         int64(item.Size),
					storageClass: item.StorageClass,
				},
				isDir: false,
			}
		}

		if !all || !response.IsTruncated {
			if objectKeyIsEmptyDir {
				o.objectsChan <- *objectKeyDetail
			}
			o.objectsChan <- listFileResult{
				endInfo: &listEndInfo{
					isTruncated: response.IsTruncated,
					nextMarker:  response.NextMarker,
				},
				ended: true,
			}
			goto END
		} else {
			marker = response.NextMarker
		}
	}
END:
}

// Get next object or pre from object lists.
func (o *objectListIterator) next() (*listFileResult, error) {
	select {
	case objectListVal := <-o.objectsChan:
		if objectListVal.err != nil {
			return nil, objectListVal.err
		}
		return &objectListVal, nil
	case <-time.After(time.Duration(bce.DEFAULT_CONNECTION_TIMEOUT_IN_MILLIS) * time.Millisecond):
		return nil, fmt.Errorf("Get Object list time out!")
	}
	return nil, fmt.Errorf("Unknown Error!")
}

func multiUploadNeedRetry(err error) bool {
	if err == nil {
		return false
	}

	// This upload id might have been aborted or completed.
	serverErr, ok := err.(*bce.BceServiceError)
	if !ok {
		return false
	}

	if serverErr.Code == boscmd.CODE_NO_SUCH_UPLOAD ||
		serverErr.Code == boscmd.CODE_INVALID_PART ||
		serverErr.Code == boscmd.CODE_INVALID_PART_ORDER {
		return true
	}
	return false
}

// Delete whole directory
// Return:
//       number of success deleted objects
//       error infomation
func (h *cliHandler) multiDeleteDir(bosClient bosClientInterface, bucketName,
	objectKey string) (int, error) {

	var (
		keyList             [MAX_DELETE_NUM_EACH_TIME]string
		unDelObjects        []api.DeleteObjectResult
		curIndex            int
		err                 error
		listResult          *listFileResult
		successDeleteObject bool = true
		successDeleteNum    int
	)

	objectLi := NewObjectListIterator(bosClient, nil, bucketName, objectKey, "", true, true, true,
		true, MAX_DELETE_NUM_EACH_TIME)

	for {
		listResult, err = objectLi.next()
		if err != nil {
			goto END
		}
		if listResult.ended {
			break
		}

		keyList[curIndex] = listResult.file.path
		curIndex += 1

		if curIndex == MAX_DELETE_NUM_EACH_TIME {
			unDelObjects, err = h.multiDeleteObjectsWithRetry(bosClient,
				keyList[:curIndex], bucketName)
			if err != nil {
				goto END
			}
			lenUnDelObjects := len(unDelObjects)
			if lenUnDelObjects != 0 {
				successDeleteObject = false
			}
			successDeleteNum += (curIndex - lenUnDelObjects)
			h.printDelResult(bucketName, keyList[:curIndex], unDelObjects)
			curIndex = 0
		}
	}

	if curIndex != 0 {
		unDelObjects, err = h.multiDeleteObjectsWithRetry(bosClient, keyList[:curIndex],
			bucketName)
		if err != nil {
			goto END
		}
		lenUnDelObjects := len(unDelObjects)
		if lenUnDelObjects != 0 {
			successDeleteObject = false
		}
		successDeleteNum += (curIndex - lenUnDelObjects)
		h.printDelResult(bucketName, keyList[:curIndex], unDelObjects)
	}

END:
	if err != nil {
		return successDeleteNum, err
	}
	if successDeleteObject {
		return successDeleteNum, nil
	}
	return successDeleteNum, fmt.Errorf("Failed to delete some objects")
}

// Delete objects, retry twice.
func (h *cliHandler) multiDeleteObjectsWithRetry(bosClient bosClientInterface, objectList []string,
	bucketName string) ([]api.DeleteObjectResult, error) {

	var (
		unDelList         []api.DeleteObjectResult
		unDelObjectsRetry *api.DeleteMultipleObjectsResult
		unDelObjects      *api.DeleteMultipleObjectsResult
		err               error
	)

	log.Infof("start multi delete objects: %s", objectList)
	unDelObjects, err = bosClient.DeleteMultipleObjectsFromKeyList(bucketName, objectList)
	// when delete objects success, gosdk may return io.EOF
	if err == io.EOF {
		err = nil
	}
	if err != nil || unDelObjects == nil {
		return nil, err
	}

	// retry to delete undeleted objects
	unDelObjectLen := len(unDelObjects.Errors)

	if unDelObjectLen == 0 {
		return nil, err
	}

	// get failed object list
	log.Debugf("Failed to delete:\n")
	keyList := make([]string, unDelObjectLen)
	curIndex := 0
	for _, unDelObject := range unDelObjects.Errors {
		if unDelObject.Code != boscmd.CODE_NO_SUCH_KEY {
			keyList[curIndex] = unDelObject.Key
			curIndex += 1
		} else {
			log.Debugf("    %s, code: %s, Message: %s", unDelObject.Key, unDelObject.Code,
				unDelObject.Message)
			unDelList = append(unDelList, unDelObject)
		}
	}

	// retry delete objects
	if curIndex > 0 {
		unDelObjectsRetry, err = bosClient.DeleteMultipleObjectsFromKeyList(bucketName,
			keyList[:curIndex])
		// when delete objects success, gosdk may return io.EOF
		if err == io.EOF {
			err = nil
		}
		if err != nil {
			unDelList = unDelObjects.Errors
		} else if unDelObjectsRetry != nil {
			for _, unDelObject := range unDelObjectsRetry.Errors {
				log.Debugf("    %s, code: %s, Message: %s", unDelObject.Key, unDelObject.Code,
					unDelObject.Message)
				unDelList = append(unDelList, unDelObject)
			}
		}
	}
	return unDelList, err
}

// delete single object
func (h *cliHandler) utilDeleteObject(bosClient bosClientInterface, bucketName,
	objectKey string) error {

	err := bosClient.DeleteObject(bucketName, objectKey)
	if err == nil {
		printIfNotQuiet("Delete object: %s%s/%s\n", BOS_PATH_PREFIX, bucketName, objectKey)
	}
	return err
}

// copy single object
func (h *cliHandler) utilCopyObject(srcBosClient, bosClient bosClientInterface, srcBucketName,
	srcObjectKey, dstBucketName, dstObjectKey, storageClass string, fileSize, fileMtime,
	timeOfgetObjectInfo int64, restart bool) error {

	var (
		err error
	)

	if fileSize > MULTI_COPY_THRESHOLD {
		// multi copy
		err = h.CopySuperFile(srcBosClient, bosClient, srcBucketName, srcObjectKey, dstBucketName,
			dstObjectKey, storageClass, fileSize, fileMtime, timeOfgetObjectInfo, restart,
			"Copying")
		// retry?
		if err != nil && multiUploadNeedRetry(err) {
			// this upload id might have been aborted or completed, so, retry and restart!
			err = h.CopySuperFile(srcBosClient, bosClient, srcBucketName, srcObjectKey,
				dstBucketName, dstObjectKey, storageClass, fileSize, fileMtime,
				timeOfgetObjectInfo, true, "Retry Copying")
		}
	} else {
		// common copy
		args := new(api.CopyObjectArgs)
		args.StorageClass = storageClass
		_, err = bosClient.CopyObject(dstBucketName, dstObjectKey, srcBucketName, srcObjectKey, args)
	}

	if err == nil {
		printIfNotQuiet("Copy: %s%s/%s to %s%s/%s\n", BOS_PATH_PREFIX, srcBucketName, srcObjectKey,
			BOS_PATH_PREFIX, dstBucketName, dstObjectKey)
	}
	return err
}

// download an object to local
func (h *cliHandler) utilDownloadObject(bosClient bosClientInterface, srcBucketName, srcObjectKey,
	dstFilePath, downLoadTmp string, yes bool, fileSize, mtime, timeOfgetObjectInfo int64,
	restart bool) error {

	var (
		dstPathEndwithSep bool
		finalFileName     string
		err               error
	)

	srcObjectName := getObjectNameFromObjectKey(srcObjectKey)
	if srcObjectName == "" {
		return fmt.Errorf("Object name error %s", srcObjectKey)
	}

	dstFilePath = replaceToOsPath(dstFilePath)
	if strings.HasSuffix(dstFilePath, util.OsPathSeparator) {
		dstPathEndwithSep = true
	}

	absFileName, err := util.Abs(dstFilePath)
	if err != nil {
		return err
	}

	// generate final file name
	if util.DoesPathExist(absFileName) {
		// TODO bug: when an object have the same name with local directory
		if util.DoesDirExist(absFileName) {
			finalFileName = filepath.Join(absFileName, srcObjectName)
		} else if dstPathEndwithSep {
			return fmt.Errorf("Can't download file, because file %s exists", absFileName)
		} else {
			finalFileName = absFileName
		}
	} else {
		if dstPathEndwithSep {
			err = util.TryMkdir(absFileName)
			if err != nil {
				return err
			}
			finalFileName = filepath.Join(absFileName, srcObjectName)
		} else {
			splitPath, _ := splitPathAndFile(absFileName)
			err = util.TryMkdir(splitPath)
			if err != nil {
				return err
			}
			finalFileName = absFileName
		}
	}

	// check whether need cover local file
	if util.DoesFileExist(finalFileName) {
		if !yes {
			yes = util.PromptConfirm("Will you cover the existing file %s?", finalFileName)
		}
		if !yes {
			return fmt.Errorf("Download abort for existing file.")
		} else if !util.IsFileWritable(finalFileName) {
			return fmt.Errorf("Download abort for covering on a existing file not writeable.")
		}
	}

	// need to get the basic information of this file
	if fileSize == 0 && mtime == 0 {
		ret, err := getObjectMeta(bosClient, srcBucketName, srcObjectKey)
		if err != nil {
			return err
		}
		fileSize = ret.size
		mtime = ret.mtime
		timeOfgetObjectInfo = ret.gtime
	}

	// start to download object to local
	if fileSize < MULTI_DOWNLOAD_THRESHOLD {
		// download small file
		err = bosClient.BasicGetObjectToFile(srcBucketName, srcObjectKey, finalFileName)
	} else {
		// download super file
		err = h.DownloadSuperFile(bosClient, srcBucketName, srcObjectKey, finalFileName,
			"Downloading", downLoadTmp, fileSize, mtime, timeOfgetObjectInfo, restart)
	}
	if err != nil {
		return err
	}
	printIfNotQuiet("Download: %s%s/%s to %s\n", BOS_PATH_PREFIX, srcBucketName, srcObjectKey,
		finalFileName)
	return nil
}

func (h *cliHandler) DownloadSuperFile(bosClient bosClientInterface, srcBucketName, srcObjectKey,
	fileName, testPrefix, downLoadTmp string, fileSize, mtime, timeOfgetObjectInfo int64,
	restart bool) (err error) {

	var (
		content *MultiTaskContent
	)
	defer func() {
		if content != nil {
			content.Flush()
			if err := util.GFinisher.Remove(content); err != nil {
				log.Debugf("can't delete content %s, when download superfile bos:/%s/%s to %s ",
					content.contentId, srcBucketName, srcObjectKey, fileName)
			}
		}
	}()

	// get object info again, we need to get the newest info of this object, as this
	// file may have been modified.
	if (time.Now().Unix() - timeOfgetObjectInfo) > GAP_GET_OBJECT_INFO_AGAIN {
		log.Debugf("need get object info again, bos:/%s/%s => %s ", srcBucketName,
			srcObjectKey, fileName)
		ret, err := getObjectMeta(bosClient, srcBucketName, srcObjectKey)
		if err != nil {
			return err
		}
		fileSize = ret.size
		mtime = ret.mtime
		timeOfgetObjectInfo = ret.gtime
	}

	// this file is samll file
	if fileSize < MULTI_DOWNLOAD_THRESHOLD {
		return bosClient.BasicGetObjectToFile(srcBucketName, srcObjectKey, fileName)
	}

	// get md5 of src object
	ranges := []int64{0, 1023, fileSize - 1025, fileSize - 1}
	md5Val, err := h.GetObjectMd5(bosClient, srcBucketName, srcObjectKey, ranges)
	if err != nil {
		return err
	}

	// get multi download part size
	multiDownloadPartSize, ok := bceconf.ServerConfigProvider.GetMultiUploadPartSize()
	if !ok {
		return fmt.Errorf("There is no info about multi download part size found!")
	}
	multiDownloadPartSizeByte := multiDownloadPartSize * (1 << 20)

	// init object content for breakpoint
	content = &MultiTaskContent{}
	err = content.init(srcBucketName, srcObjectKey, "", fileName, IS_BOS, IS_LOCAL,
		md5Val, fileSize, mtime, restart, multiDownloadPartSizeByte)
	if err != nil {
		return err
	}

	util.GFinisher.Insert(content)

	// Do the parallel multipart download
	if content.needRestart {
		// UploadId is the name of temporary file that stores the intermediate Data
		if downLoadTmp != "" {
			if util.DoesFileExist(downLoadTmp) {
				return fmt.Errorf("%s is a file, it should be a directory !")
			} else if !util.DoesDirExist(downLoadTmp) {
				return fmt.Errorf("Temporary folder %s don't exist!", downLoadTmp)
			}
			content.uploadId = filepath.Join(downLoadTmp, "bcecmd.temp."+content.contentId)
		} else {
			content.uploadId = filepath.Dir(fileName) + util.OsPathSeparator +
				"bcecmd.temp." + content.contentId
		}
	}

	// for progress bar
	bar, err := util.NewBar(int(content.partsNum+1), testPrefix, Quiet || DisableBar)
	if err != nil {
		return err
	}
	util.GFinisher.Insert(bar)
	defer func() {
		bar.Exit()
		util.GFinisher.Remove(bar)
	}()
	bar.Finish(content.GetFinshPartNum())

	// temp file for save intermediate result
	file, err := os.OpenFile(content.uploadId, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	log.Debugf("starting download super file, total parts: %d, part size: %d", content.partsNum,
		content.partSize)

	// TODO: use diff thread number with  multi-upload
	multiDownloadThreadNum, ok := bceconf.ServerConfigProvider.GetMultiUploadThreadNum()
	if !ok {
		return fmt.Errorf("There is no info about multi download thread Num found!")
	}

	downloadPart := func(partId int64, rangeStart, rangeEnd, workerId int64, ret chan error,
		pool chan int64, doneChan chan struct{}) {
		res, rangeGetErr := bosClient.GetObject(srcBucketName, srcObjectKey, nil, rangeStart,
			rangeEnd)
		if rangeGetErr != nil {
			log.Errorf("download object part(offset:%d, size:%d) failed: %v",
				rangeStart, res.ContentLength, rangeGetErr)
			ret <- rangeGetErr
			return
		}
		defer res.Body.Close()
		log.Debugf("%s writing part %d with offset=%d, size=%d", fileName, partId, rangeStart,
			res.ContentLength)
		buf := make([]byte, 1048576)
		offset := rangeStart
		for {
			n, e := res.Body.Read(buf)
			if e != nil && e != io.EOF {
				ret <- e
				return
			}
			if n == 0 {
				break
			}
			if _, writeErr := file.WriteAt(buf[:n], offset); writeErr != nil {
				ret <- writeErr
				return
			}
			offset += int64(n)
		}
		log.Debugf("%s writing part %d done", fileName, partId)
		content.finishPart(partId, strconv.FormatInt(partId, 10))
		bar.Finish(content.GetFinshPartNum())
		pool <- workerId
		doneChan <- struct{}{}
	}

	afterDownPartFail := func(err error) {
		serverErr, ok := err.(*bce.BceServiceError)
		if !ok {
			return
		}
		if serverErr.Code == boscmd.CODE_NO_SUCH_KEY && file != nil {
			file.Close()
			content.Remove()
		}
	}

	// Set up multiple goroutine workers to download the object
	doneChan := make(chan struct{}, content.partsNum)
	retChan := make(chan error)
	workerPool := make(chan int64, multiDownloadThreadNum)

	for i := int64(0); i < multiDownloadThreadNum; i++ {
		workerPool <- i
	}

	log.Debugf("content partsNum:%d", content.partsNum)
	for partId := int64(1); partId <= content.partsNum; partId++ {
		if _, ok := content.partIsFinish(partId); ok {
			doneChan <- struct{}{}
			log.Debugf("skipping part %d, because it has been downloaded copy bos:/%s/%s => "+
				"%s", partId, srcBucketName, srcObjectKey, fileName)
			continue
		}
		rangeStart := (partId - 1) * content.partSize
		rangeEnd := partId * content.partSize
		if rangeEnd > fileSize {
			rangeEnd = fileSize
		}
		rangeEnd--

		select {
		case workerId := <-workerPool:
			log.Debugf("download part partid:%d", partId)
			go downloadPart(partId, rangeStart, rangeEnd, workerId, retChan, workerPool, doneChan)
		case downloadErr := <-retChan:
			afterDownPartFail(downloadErr)
			return downloadErr
		}
	}

	// Wait for writing to local file done
	for i := content.partsNum; i > 0; i-- {
		select {
		case <-doneChan:
			continue
		case downloadErr := <-retChan:
			afterDownPartFail(downloadErr)
			return downloadErr
		}
	}

	// fail to close file, does need to remove temp file ?
	if err := file.Close(); err != nil {
		return err
	}

	if err := os.Rename(content.uploadId, fileName); err != nil {
		return err
	}
	file = nil
	bar.Finish(content.GetFinshPartNum() + 1)
	content.complete()
	return nil
}

// upload a file
func (h *cliHandler) utilUploadFile(bosClient bosClientInterface, srcPath, relSrcPath,
	dstBucketName, dstObjectKey, storageClass string, fileSize, fileMtime,
	timeOfgetObjectInfo int64, restart bool) error {

	var (
		err error
	)

	if fileSize > MULTI_UPLOAD_THRESHOLD {
		err = h.UploadSuperFile(bosClient, relSrcPath, dstBucketName, dstObjectKey, storageClass,
			fileSize, fileMtime, timeOfgetObjectInfo, restart, "Uploading")
		if err != nil && multiUploadNeedRetry(err) {
			//this upload id might have been aborted or completed, so, retry and restart!
			err = h.UploadSuperFile(bosClient, relSrcPath, dstBucketName, dstObjectKey,
				storageClass, fileSize, fileMtime, timeOfgetObjectInfo, true, "Retry Uploading")
		}
	} else {
		// TODO putObject of go sdk don't have interface for storage-class
		args := &api.PutObjectArgs{StorageClass: storageClass}
		_, err = bosClient.PutObjectFromFile(dstBucketName, dstObjectKey, relSrcPath, args)
	}

	if err != nil {
		return err
	}
	printIfNotQuiet("Upload: %s to %s%s/%s\n", srcPath, BOS_PATH_PREFIX, dstBucketName,
		dstObjectKey)
	return nil
}

// UploadSuperFile - parallel upload the super file by using the multipart upload interface
func (h *cliHandler) UploadSuperFile(bosClient bosClientInterface, srcPath, dstBucketName,
	dstObjectKey, storageClass string, fileSize, mtime, timeOfgetObjectInfo int64,
	restart bool, testPrefix string) error {

	var (
		content *MultiTaskContent
	)

	defer func() {
		if content != nil {
			// flush for normal exit
			content.Flush()
			if err := util.GFinisher.Remove(content); err != nil {
				log.Debugf("can't delete content %s, $s => bos:/%s/%s ", content.contentId,
					srcPath, dstBucketName, dstObjectKey)
			}
		}
	}()

	// get object info again, we need to get the newest info of this object, as this
	// file may have been modified.
	if (time.Now().Unix() - timeOfgetObjectInfo) > GAP_GET_OBJECT_INFO_AGAIN {
		log.Debugf("need get object info again, %s => bos:/%s/%s ", srcPath,
			dstBucketName, dstObjectKey)
		ret, err := getFileMate(srcPath)
		if err != nil {
			return err
		}
		fileSize = ret.size
		mtime = ret.mtime
		timeOfgetObjectInfo = ret.gtime
	}

	if fileSize < MULTI_UPLOAD_THRESHOLD {
		args := &api.PutObjectArgs{StorageClass: storageClass}
		_, err := bosClient.PutObjectFromFile(dstBucketName, dstObjectKey, srcPath, args)
		return err
	}

	// open file for read
	fd, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer fd.Close()

	// get md5 of src object
	md5Val, err := util.GetFileMd5(fd, -65536, 65536, 2)
	if err != nil {
		return err
	}

	// get multi upload part size
	multiUploadPartSize, ok := bceconf.ServerConfigProvider.GetMultiUploadPartSize()
	if !ok {
		return fmt.Errorf("There is no info about multi upload part size found!")
	}
	multiUploadPartSizeByte := multiUploadPartSize * (1 << 20)

	// init object content for breakpoint
	content = &MultiTaskContent{}
	err = content.init("", srcPath, dstBucketName, dstObjectKey, IS_LOCAL, IS_BOS,
		md5Val, fileSize, mtime, restart, multiUploadPartSizeByte)
	if err != nil {
		return err
	}

	util.GFinisher.Insert(content)

	// Do the parallel multipart upload
	if content.needRestart {
		resp, err := bosClient.InitiateMultipartUpload(dstBucketName, dstObjectKey, "",
			&api.InitiateMultipartUploadArgs{StorageClass: storageClass})
		if err != nil {
			return err
		}
		content.uploadId = resp.UploadId
	}

	// for progress bar
	bar, err := util.NewBar(int(content.partsNum+1), testPrefix, Quiet || DisableBar)
	if err != nil {
		return err
	}

	util.GFinisher.Insert(bar)
	defer func() {
		bar.Exit()
		util.GFinisher.Remove(bar)
	}()
	bar.Finish(content.GetFinshPartNum())

	// Inner wrapper function of parallel uploading each part to get the ETag of the part
	uploadPart := func(bucket, object, uploadId string, partNumber int64, partBody []byte,
		result chan *api.UploadInfoType, ret chan error, id int64, pool chan int64) {
		log.Debugf("%s => bos:/%s/%s start upload partNumber %d\n", srcPath, dstBucketName,
			dstObjectKey, partNumber)

		etag, err := bosClient.UploadPartFromBytes(bucket, object, uploadId, int(partNumber),
			partBody, nil)

		if err != nil {
			log.Debugf("finish upload part %d from %s => %s, error is %s",
				partNumber, srcPath, dstBucketName, dstObjectKey, err)
			result <- nil
			ret <- err
		} else if etag == "" {
			log.Debugf("failed upload part %d from %s => %s, because we get a empty etag",
				partNumber, srcPath, dstBucketName, dstObjectKey)
			result <- nil
			ret <- fmt.Errorf("get a empty etag when upload part %d", partNumber)
		} else {
			if finErr := content.finishPart(partNumber, etag); err != nil {
				result <- nil
				ret <- finErr
			} else {
				bar.Finish(content.GetFinshPartNum())
				log.Debugf("finish upload part %d from %s => bos:/%s/%s, etag is %s",
					partNumber, srcPath, dstBucketName, dstObjectKey, etag)
				result <- &api.UploadInfoType{int(partNumber), etag}
			}
		}
		pool <- id
	}

	// get multi upload or copy thread num
	multiUploadThreadNum, ok := bceconf.ServerConfigProvider.GetMultiUploadThreadNum()
	if !ok {
		return fmt.Errorf("There is no info about multi upload thread Num found!")
	}

	uploadedResult := make(chan *api.UploadInfoType, content.partsNum)
	retChan := make(chan error, content.partsNum)
	workerPool := make(chan int64, multiUploadThreadNum)
	for i := int64(0); i < multiUploadThreadNum; i++ {
		workerPool <- i
	}

	for partId := int64(1); partId <= content.partsNum; partId++ {
		if partInfo, ok := content.partIsFinish(partId); ok {
			uploadedResult <- &api.UploadInfoType{int(partInfo.PartNumberId), partInfo.ETag}
			log.Debugf("skipping part %d, because it has been uploaded %s => "+
				"bos:/%s/%s", partId, srcPath, dstBucketName, dstObjectKey)
			continue
		}

		offset := (partId - 1) * content.partSize
		uploadSize := content.partSize
		if offset+uploadSize > fileSize {
			uploadSize = fileSize - offset
		}

		// read part from file to bytes
		partBody := make([]byte, uploadSize, uploadSize)
		n, err := fd.ReadAt(partBody, offset)
		if err != nil {
			return err
		} else if int64(n) != uploadSize {
			return fmt.Errorf("read size %d != upload size %d!", n, uploadSize)
		}

		select { // wait until get a worker to upload
		case workerId := <-workerPool:
			go uploadPart(dstBucketName, dstObjectKey, content.uploadId, partId, partBody,
				uploadedResult, retChan, workerId, workerPool)
		case uploadPartErr := <-retChan:
			return uploadPartErr
		}
	}

	// Check the return of each part uploading, and decide to complete or abort it
	completeArgs := &api.CompleteMultipartUploadArgs{
		Parts: make([]api.UploadInfoType, content.partsNum),
	}
	for i := content.partsNum; i > 0; i-- {
		uploaded := <-uploadedResult
		if uploaded == nil { // error occurs and not be caught in `select' statement
			return <-retChan
		} else if uploaded.ETag == "" {
			return fmt.Errorf("get empty etag when upload part %d", uploaded.PartNumber)
		}
		completeArgs.Parts[uploaded.PartNumber-1] = *uploaded
		log.Debugf("success copy upload part %d from %s => bos:/%s/%s , etag is %s",
			srcPath, dstBucketName, dstObjectKey, uploaded.PartNumber, uploaded.ETag)
	}

	if _, err := bosClient.CompleteMultipartUploadFromStruct(dstBucketName, dstObjectKey,
		content.uploadId, completeArgs); err != nil {
		return err
	} else {
		bar.Finish(content.GetFinshPartNum() + 1)
		content.complete()
	}

	return nil
}

// delete local file
func (h *cliHandler) utilDeleteLocalFile(localPath string) error {
	if util.DoesDirExist(localPath) {
		return fmt.Errorf("%s is directory", localPath)
	}
	err := os.Remove(localPath)
	if err == nil {
		printIfNotQuiet("Delete file: %s\n", localPath)
	}
	return err
}

// check whether there is a bucket with specific name
func (h *cliHandler) doesBucketExist(bosClient bosClientInterface, bucketName string) (bool,
	error) {

	_, ok := boscmd.GetEndpointOfBucketFromCache(bucketName)
	if ok {
		return true, nil
	}
	err := bosClient.HeadBucket(bucketName)
	if err == nil {
		return true, nil
	}
	if serverErr, ok := err.(*bce.BceServiceError); ok {
		// TODO if bucket don't exist, HeadBucket will return wrong error message, waiting go sdk
		//fix this bug
		if serverErr.StatusCode == http.StatusNotFound {
			return false, nil
		} else if serverErr.StatusCode == http.StatusForbidden {
			return true, nil
		}
	}
	return false, err
}

// Compare success list and filed list,  print delete result
func (h *cliHandler) printDelResult(bucketName string, keyList []string,
	unDelObjects []api.DeleteObjectResult) {

	var (
		keyListIndex int
		delListIndex int
	)

	if Quiet {
		return
	}

	lenKeyList := len(keyList)
	lenDelList := len(unDelObjects)

	unDelList := make([]string, lenDelList)

	for i, unDelObject := range unDelObjects {
		unDelList[i] = unDelObject.Key
	}
	sort.Strings(unDelList)

	for {
		if keyListIndex < lenKeyList && delListIndex < lenDelList {
			if keyList[keyListIndex] == unDelList[delListIndex] {
				printIfNotQuiet("Failed delete object: %s%s/%s\n", BOS_PATH_PREFIX, bucketName,
					keyList[keyListIndex])
				delListIndex += 1
				keyListIndex += 1
			} else if keyList[keyListIndex] < unDelList[delListIndex] {
				printIfNotQuiet("Delete object: %s%s/%s\n", BOS_PATH_PREFIX, bucketName,
					keyList[keyListIndex])
				keyListIndex += 1
			} else {
				// TODO TEST print info
				delListIndex += 1
				// bcecliAbnormalExistMsg("**undeleted objects list is not sorted **")
			}
		} else if keyListIndex < lenKeyList {
			printIfNotQuiet("Delete object: %s%s/%s\n", BOS_PATH_PREFIX, bucketName,
				keyList[keyListIndex])
			keyListIndex += 1
		} else if delListIndex < lenDelList {
			fmt.Printf("Del list: '%s'\n", unDelList[delListIndex])
			delListIndex += 1
		} else {
			break
		}
	}
}

// calc the md5 of parts of object
func (h *cliHandler) GetObjectMd5(bosClient bosClientInterface, bucketName, objectKey string,
	ranges []int64) (string, error) {

	rangesLen := len(ranges)
	md5New := md5.New()

	if rangesLen == 0 || rangesLen%2 != 0 {
		return "", fmt.Errorf("Invalid ranges")
	}
	for i := 0; i < rangesLen; i += 2 {
		ret, err := bosClient.GetObject(bucketName, objectKey, nil, ranges[i], ranges[i+1])
		if err != nil {
			return "", err
		}
		needCopied := ranges[i+1] - ranges[i] + 1
		copied, err := io.Copy(md5New, ret.Body)
		if err != nil {
			return "", err
		} else if copied != needCopied {
			return "", fmt.Errorf("get %d bytes intead %d from %s where start is %d and end is %d",
				copied, needCopied, BOS_PATH_PREFIX+bucketName+util.BOS_PATH_SEPARATOR+objectKey,
				ranges[i], ranges[i+1])
		}
	}
	return hex.EncodeToString(md5New.Sum(nil)), nil
}

// CopySuperFile - parallel upload the super file by using the multipart upload interface
func (h *cliHandler) CopySuperFile(srcBosClient, bosClient bosClientInterface, srcBucketName,
	srcObjectKey, dstBucketName, dstObjectKey, storageClass string, fileSize, mtime,
	timeOfgetObjectInfo int64, restart bool, testPrefix string) error {

	var (
		content *MultiTaskContent
	)
	defer func() {
		if content != nil {
			content.Flush()
			if err := util.GFinisher.Remove(content); err != nil {
				log.Debugf("can't delete content %s, bos:/%s/%s => bos:/%s/%s ", content.contentId,
					srcBucketName, srcObjectKey, dstBucketName, dstObjectKey)
			}
		}
	}()

	// get object info again, we need to get the newest info of this object, as this
	// file may have been modified.
	if (time.Now().Unix() - timeOfgetObjectInfo) > GAP_GET_OBJECT_INFO_AGAIN {
		log.Debugf("need get object info again, bos:/%s/%s => bos:/%s/%s ", srcBucketName,
			srcObjectKey, dstBucketName, dstObjectKey)
		ret, err := getObjectMeta(srcBosClient, srcBucketName, srcObjectKey)
		if err != nil {
			return err
		}
		fileSize = ret.size
		mtime = ret.mtime
		timeOfgetObjectInfo = ret.gtime
	}

	if fileSize < MULTI_COPY_THRESHOLD {
		args := new(api.CopyObjectArgs)
		args.StorageClass = storageClass
		_, err := bosClient.CopyObject(dstBucketName, dstObjectKey, srcBucketName, srcObjectKey, args)
		return err
	}

	// get md5 of src object
	ranges := []int64{0, 1023, fileSize - 1025, fileSize - 1}
	md5Val, err := h.GetObjectMd5(srcBosClient, srcBucketName, srcObjectKey, ranges)
	if err != nil {
		return err
	}

	// init object content for breakpoint
	content = &MultiTaskContent{}
	err = content.init(srcBucketName, srcObjectKey, dstBucketName, dstObjectKey, IS_BOS, IS_BOS,
		md5Val, fileSize, mtime, restart, MULTI_COPY_PART_SIZE)
	if err != nil {
		return err
	}

	util.GFinisher.Insert(content)

	// Do the parallel multipart upload
	if content.needRestart {
		resp, err := bosClient.InitiateMultipartUpload(dstBucketName, dstObjectKey, "",
			&api.InitiateMultipartUploadArgs{StorageClass: storageClass})
		if err != nil {
			return err
		}
		content.uploadId = resp.UploadId
	}

	// for progress bar
	bar, err := util.NewBar(int(content.partsNum+1), testPrefix, Quiet || DisableBar)
	if err != nil {
		return err
	}
	util.GFinisher.Insert(bar)
	defer func() {
		bar.Exit()
		util.GFinisher.Remove(bar)
	}()
	bar.Finish(content.GetFinshPartNum())

	// Inner wrapper function of parallel uploading each part to get the ETag of the part
	copyPart := func(srcBucket, srcObject, dstBucket, dstObject, uploadId string, partNumber,
		partSize, fileSize int64, result chan *api.UploadInfoType, ret chan error, id int64,
		pool chan int64) {

		rangeStart := (partNumber - 1) * partSize
		rangeEnd := partNumber * partSize
		if rangeEnd > fileSize {
			rangeEnd = fileSize
		}
		rangeEnd--

		args := &api.UploadPartCopyArgs{
			SourceRange: fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd),
		}

		log.Debugf("bos:/%s/%s => bos:/%s/%s start copy partNumber %d partSize %d range is %s , "+
			"size is %d\n", srcBucketName, srcObjectKey, dstBucketName, dstObjectKey, partNumber,
			partSize, args.SourceRange, rangeEnd-rangeStart)

		copyRet, err := bosClient.UploadPartCopy(dstBucketName, dstObjectKey, srcBucketName,
			srcObjectKey, uploadId, int(partNumber), args)

		if err != nil {
			log.Debugf("finish copy part %d from bos:/%s/%s => bos:/%s/%s, error is %s",
				partNumber, srcBucketName, srcObjectKey, dstBucketName, dstObjectKey, err)
			result <- nil
			ret <- err
		} else if copyRet.ETag == "" {
			log.Debugf("failed copy part %d from bos:/%s/%s => bos:/%s/%s, because we get a "+
				"empty etag", partNumber, srcBucketName, srcObjectKey, dstBucketName, dstObjectKey)
			result <- nil
			ret <- fmt.Errorf("get a empy etag when copy part %d", partNumber)
		} else {
			if finErr := content.finishPart(partNumber, copyRet.ETag); err != nil {
				result <- nil
				ret <- finErr
			} else {
				log.Debugf("finish copy part %d from bos:/%s/%s => bos:/%s/%s, etag is %s",
					partNumber, srcBucketName, srcObjectKey, dstBucketName, dstObjectKey,
					copyRet.ETag)
				bar.Finish(content.GetFinshPartNum())
				result <- &api.UploadInfoType{int(partNumber), copyRet.ETag}
			}
		}
		pool <- id
	}

	// get multi upload or copy thread num
	multiUploadThreadNum, ok := bceconf.ServerConfigProvider.GetMultiUploadThreadNum()
	if !ok {
		return fmt.Errorf("There is no info about multi upload thread Num found!")
	}

	uploadedResult := make(chan *api.UploadInfoType, content.partsNum)
	retChan := make(chan error, content.partsNum)
	workerPool := make(chan int64, multiUploadThreadNum)
	for i := int64(0); i < multiUploadThreadNum; i++ {
		workerPool <- i
	}

	for partId := int64(1); partId <= content.partsNum; partId++ {
		if partInfo, ok := content.partIsFinish(partId); ok {
			uploadedResult <- &api.UploadInfoType{int(partInfo.PartNumberId), partInfo.ETag}
			log.Debugf("skipping part %d, because it has been uploaded copy bos:/%s/%s => "+
				"bos:/%s/%s", partId, srcBucketName, srcObjectKey, dstBucketName, dstObjectKey)
			continue
		}
		select { // wait until get a worker to upload
		case workerId := <-workerPool:
			go copyPart(srcBucketName, srcObjectKey, dstBucketName, dstObjectKey, content.uploadId,
				partId, content.partSize, fileSize, uploadedResult, retChan, workerId, workerPool)
		case uploadPartErr := <-retChan:
			return uploadPartErr
		}
	}

	// Check the return of each part uploading, and decide to complete or abort it
	completeArgs := &api.CompleteMultipartUploadArgs{
		Parts: make([]api.UploadInfoType, content.partsNum),
	}
	for i := content.partsNum; i > 0; i-- {
		uploaded := <-uploadedResult
		if uploaded == nil { // error occurs and not be caught in `select' statement
			return <-retChan
		} else if uploaded.ETag == "" {
			return fmt.Errorf("get empty etag when upload part %d", uploaded.PartNumber)
		}
		completeArgs.Parts[uploaded.PartNumber-1] = *uploaded
		log.Debugf("success copy upload part %d from bos:/%s/%s => bos:/%s/%s , etag is %s",
			srcBucketName, srcObjectKey, dstBucketName, dstObjectKey, uploaded.PartNumber,
			uploaded.ETag)
	}

	if _, err := bosClient.CompleteMultipartUploadFromStruct(dstBucketName, dstObjectKey,
		content.uploadId, completeArgs); err != nil {
		return err
	} else {
		bar.Finish(content.GetFinshPartNum() + 1)
		content.complete()
	}

	return nil
}

func getObjectMeta(bosClient bosClientInterface, bucketName,
	objectKey string) (*fileDetail, error) {
	if bucketName == "" || objectKey == "" {
		return nil, fmt.Errorf("bucket name and object name can not be empty!")
	}

	getMetaRet, err := bosClient.GetObjectMeta(bucketName, objectKey)
	if err != nil {
		// TODO if object don't exist, getObjectMeta will return wrong error message
		if serverErr, ok := err.(*bce.BceServiceError); ok {
			if serverErr.StatusCode == 404 {
				serverErr.Code = boscmd.CODE_NO_SUCH_KEY
				serverErr.Message = "Object don't exist!"
			}
		}
		return nil, err
	}
	// utc to timestamp
	mtime, err := util.TranUTCTimeStringToTimeStamp(getMetaRet.LastModified, BOS_HTTP_TIME_FORMT)
	if err != nil {
		return nil, err
	}

	return &fileDetail{
		path:         objectKey,
		mtime:        mtime,
		gtime:        time.Now().Unix(),
		size:         getMetaRet.ContentLength,
		storageClass: getMetaRet.StorageClass,
		crc32:        getMetaRet.ContentCrc32,
	}, nil
}
