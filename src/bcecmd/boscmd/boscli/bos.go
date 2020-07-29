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

// This module provides the major operations on BOS.

package boscli

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

import (
	"bcecmd/boscmd"
	"bceconf"
	"utils/util"
)

var (
	Quiet                 bool
	DisableBar            bool // display or not display progress bar
	IsConcurrentOperation bool
)

// Create new BosCli
func NewBosCli() *BosCli {
	var (
		ak       string
		sk       string
		endpoint string
		err      error
	)

	boscliClient := &BosCli{}
	boscliClient.bosClient, err = bosClientInit(ak, sk, endpoint)
	if err != nil {
		bcecliAbnormalExistErr(err)
	}
	boscliClient.handler = &cliHandler{}
	return boscliClient
}

type BosCli struct {
	bosClient bosClientInterface
	handler   handlerInterface
}

type operateResult struct {
	successed int
	failed    int
}

type genSignedUrlArgs struct {
	bucketName string
	objectKey  string
	expires    int
}

// generate signed url for bospath
func (b *BosCli) GenSignedUrl(bosPath string, expires int, haveSetExpires bool) {
	args, retCode := b.genSignedUrlPreProcess(bosPath, expires, haveSetExpires)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// execute genSignedUrl
	bosUrl := b.bosClient.BasicGeneratePresignedUrl(args.bucketName, args.objectKey, args.expires)
	fmt.Println(bosUrl)
}

// request check and preprocessing for genSignedUrl
func (b *BosCli) genSignedUrlPreProcess(bosPath string, expires int, haveSetExpires bool) (
	*genSignedUrlArgs, BosCliErrorCode) {

	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return nil, retCode
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return nil, BOSCLI_BUCKETNAME_IS_EMPTY
	}
	if objectKey == "" {
		return nil, BOSCLI_OBJECTKEY_IS_EMPTY
	}
	if haveSetExpires {
		if expires < -1 {
			return nil, BOSCLI_EXPIRE_LESS_NONE
		}
	} else {
		expires = SIGNED_URL_EXPIRE_TIME
	}
	return &genSignedUrlArgs{
		bucketName: bucketName,
		objectKey:  objectKey,
		expires:    expires,
	}, BOSCLI_OK
}

// List buckets or objects
// param: must have BOS_PATH attribute.
func (b *BosCli) List(bosPath string, all bool, recursive bool, summary bool) {
	retCode, err := checkBosPath(bosPath)
	if err != nil {
		bcecliAbnormalExistCodeErr(retCode, err)
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		_, err = b.listBuckets(summary)
	} else {
		err = b.listObjects(bucketName, objectKey, all, recursive, summary)
	}
	if err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// implement list bucket
func (b *BosCli) listBuckets(sum bool) (int, error) {
	buckets, err := b.bosClient.ListBuckets()
	if err != nil {
		return 0, err
	}

	for _, bucket := range buckets.Buckets {
		localTime, _ := util.TranUTCtoLocalTime(bucket.CreationDate, BOS_TIME_FORMT,
			LOCAL_TIME_FROMT)
		fmt.Printf("  %s  %7s  %s\n", localTime, bucket.Location, bucket.Name)
	}
	bucketsNum := len(buckets.Buckets)
	if sum {
		fmt.Printf(" Total Buckets: %d\n", bucketsNum)
	}
	return bucketsNum, nil
}

// implement list objects
func (b *BosCli) listObjects(bucketName, objectKey string, all, recursive, summary bool) error {
	var (
		preNum     int64
		objectNum  int64
		objectSize int64
	)

	objectsList := NewObjectListIterator(b.bosClient, nil, bucketName, objectKey, "", all,
		recursive, true, false, 1000)
	for {
		listResult, err := objectsList.next()
		if err != nil {
			return err
		}
		if listResult.ended {
			if listResult.endInfo.isTruncated {
				fmt.Println("more......")
			}
			break
		}
		if listResult.isDir {
			// Print pre
			preNum++
			fmt.Printf("  %19s %11s  %15s  %s\n", "", "", "PRE", listResult.dir.key)
		} else {
			// Print objects
			object := listResult.file
			localTime := util.TranTimestamptoLocalTime(object.mtime, LOCAL_TIME_FROMT)
			fmt.Printf("  %s %15d  %11s  %s\n", localTime, object.size, object.storageClass,
				object.key)
			objectSize += int64(object.size)
			objectNum++
		}
	}
	// print summary
	if summary {
		fmt.Printf("Total PRE(s): %d\n", preNum)
		fmt.Printf("Total Object(s): %d\n", objectNum)
		fmt.Printf("Total Size Of Objects(byte): %d\n", objectSize)
	}
	return nil
}

// Make bucket
func (b *BosCli) MakeBucket(bucketName, region string, quiet bool) {
	Quiet = quiet
	// preprocessing
	bucketName, retCode := b.makeBucketPreProcess(bucketName, region)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// execute make bucket
	location, err := b.bosClient.PutBucket(bucketName)
	if err != nil {
		bcecliAbnormalExistErr(err)
	}

	// print resut
	printIfNotQuiet("Make bucket: %s in region %s\n", bucketName, location)
}

// preprocessing and check make bucket request
func (b *BosCli) makeBucketPreProcess(bucketName, region string) (string, BosCliErrorCode) {
	retCode, err := checkBosPath(bucketName)
	if err != nil {
		return "", retCode
	}

	bucketName, objectKey := splitBosBucketKey(bucketName)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}

	defaultRegion, _ := bceconf.ServerConfigProvider.GetRegion()
	if region == "" || region == defaultRegion {
		return bucketName, BOSCLI_OK
	}

	// if auto switch is enabled, modifiy the endpoint according the val of region
	if autoSwitch, _ := bceconf.ServerConfigProvider.GetUseAutoSwitchDomain(); autoSwitch {
		endpoint, _ := bceconf.ServerConfigProvider.GetDomainByRegion(region)
		clientWrapper, ok := b.bosClient.(*bosClientWrapper)
		if !ok {
			return "", BOSCLI_INTERNAL_ERROR
		}
		modifiyBosClientEndpointByEndpoint(clientWrapper.bosClient, endpoint)
	} else {
		ignoreRegion := util.PromptConfirm("Option 'auto switch domain' is turned off,"+
			"default region %s will be used instead of %s, do you want to continue?",
			defaultRegion, region)
		if !ignoreRegion {
			return "", BOSCLI_OPRATION_CANCEL
		}
	}
	return bucketName, BOSCLI_OK
}

// rb: Command remove bucket from bos.
// param must have BUCKET_NAME
// force: delete bucket and ALL objects in it.
// yes: delete bucket without any prompt
func (b *BosCli) RemoveBucket(bucketName string, force, yes, quiet bool) {
	Quiet = quiet

	// preprocessing
	bucket, retCode := b.removeBucketPreProcess(bucketName)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// execute remove bucket
	retCode, err := b.removeBucketExecute(bucket, yes, force)
	if err != nil {
		bcecliAbnormalExistCodeErr(retCode, err)
	} else {
		bceconf.BucketEndpointCacheProvider.Delete(bucket)
	}

	// print resut
	printIfNotQuiet("Remove bucket: %s\n", bucket)
}

// check request of remove bucket
func (b *BosCli) removeBucketPreProcess(bucketName string) (string, BosCliErrorCode) {
	if bucketName == "" || strings.HasPrefix(bucketName, boscmd.BOS_PATH_SEPARATOR) {
		return "", boscmd.CODE_INVALID_BUCKET_NAME
	}

	bucketName, objectKey := splitBosBucketKey(bucketName)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	}
	if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}
	return bucketName, BOSCLI_OK
}

// implement remove bucket
func (b *BosCli) removeBucketExecute(bucketName string, yes, force bool) (BosCliErrorCode, error) {
	var err error

	confirmed := yes
	if force {
		if !confirmed {
			confirmed = util.PromptConfirm("Do you really want to REMOVE bucket %s%s and all "+
				"objects in it?", BOS_PATH_PREFIX, bucketName)
		}
		if confirmed {
			_, err = b.handler.multiDeleteDir(b.bosClient, bucketName, "")
			if err != nil {
				return BOSCLI_EMPTY_CODE, err
			}
			err = b.bosClient.DeleteBucket(bucketName)
			if err != nil {
				return BOSCLI_EMPTY_CODE, err
			}
		} else {
			return BOSCLI_OPRATION_CANCEL, fmt.Errorf("")
		}
	} else {
		if !confirmed {
			confirmed = util.PromptConfirm("Do you really want to REMOVE bucket %s%s?",
				BOS_PATH_PREFIX, bucketName)
		}
		if confirmed {
			err = b.bosClient.DeleteBucket(bucketName)

			if err != nil {
				return BOSCLI_EMPTY_CODE, err
			}
		} else {
			return BOSCLI_OPRATION_CANCEL, fmt.Errorf("")
		}
	}
	return BOSCLI_OK, err
}

type removeObjectArgs struct {
	objectKey  string
	bucketName string
	isDir      bool
}

// rm: remove object from bucekt, must have bos_path
// PARAMS:
//   bosPath   : bos path
//   yes       : delete object without prompts.
//   recursive : delete objects under subdirs.
//   quit      : do not display the operations performed from the specified command
func (b *BosCli) RemoveObject(bosPath string, yes, recursive, quiet bool) {
	Quiet = quiet
	// preprocessing and check request
	args, retCode := b.removeObjectPreProcess(bosPath, recursive)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// execute remove
	removed, err := b.removeObjectExecute(args, yes)

	if !(removed == 1 && IsConcurrentOperation) {
		printIfNotQuiet("[%d] objects removed on remote.\n", removed)
	}
	if err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// preprocessing and check remove object request
func (b *BosCli) removeObjectPreProcess(bosPath string, recursive bool) (*removeObjectArgs,
	BosCliErrorCode) {

	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return nil, retCode
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return nil, BOSCLI_BUCKETNAME_IS_EMPTY
	}

	if !recursive {
		if objectKey == "" || strings.HasSuffix(objectKey, boscmd.BOS_PATH_SEPARATOR) {
			return nil, BOSCLI_RM_DIR_MUST_USE_RECURSIVE
		}
		return &removeObjectArgs{
			bucketName: bucketName,
			objectKey:  objectKey,
		}, BOSCLI_OK
	}
	return &removeObjectArgs{
		bucketName: bucketName,
		objectKey:  objectKey,
		isDir:      true,
	}, BOSCLI_OK
}

// execute remove object or remove obejcts
func (b *BosCli) removeObjectExecute(args *removeObjectArgs, yes bool) (int, error) {
	var (
		deleted int
		err     error
	)

	// recursive delete
	if args.isDir {
		if !yes {
			yes = util.PromptConfirm("Do you really want to DELETE all objects in %s%s/%s?",
				BOS_PATH_PREFIX, args.bucketName, args.objectKey)
		}
		if yes {
			deleted, err = b.handler.multiDeleteDir(b.bosClient, args.bucketName, args.objectKey)
		}
		goto END
	}

	// delete single object
	if !yes {
		yes = util.PromptConfirm("Do you really want to DELETE object %s%s/%s?", BOS_PATH_PREFIX,
			args.bucketName, args.objectKey)
	}
	if yes {
		if err = b.handler.utilDeleteObject(b.bosClient, args.bucketName,
			args.objectKey); err == nil {
			deleted = 1
		}
	}

END:
	return deleted, err
}

// cp : upload, download or copy
// param args: Parsed args, must have SRC, DST, force, no_override
// exception: Both SRC and DST are local path or stream
func (b *BosCli) Copy(srcPath, dstPath, storageClass, downLoadTmp string, recursive, restart, quiet,
	yes, disableBar bool) {

	var (
		retCode BosCliErrorCode
		err     error
	)

	Quiet = quiet
	DisableBar = disableBar

	isSourceRemotePath := strings.HasPrefix(srcPath, BOS_PATH_PREFIX)
	isDestinationRemotePath := strings.HasPrefix(dstPath, BOS_PATH_PREFIX)

	if isSourceRemotePath && isDestinationRemotePath {
		retCode, err = b.copyBetweenRemote(srcPath, dstPath, storageClass, recursive, restart)
	} else if isSourceRemotePath {
		retCode, err = b.copyDownload(srcPath, dstPath, downLoadTmp, recursive, yes, restart)
	} else if isDestinationRemotePath {
		retCode, err = b.copyUpload(srcPath, dstPath, storageClass, recursive, restart)
	} else {
		bcecliAbnormalExistMsg("You can use cp/copy to copy files between local file system.")
	}
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCodeErr(retCode, err)
	}
}

type copyBetweenRemoteArgs struct {
	srcBucketName string
	srcObjectKey  string
	dstBucketName string
	dstObjectKey  string
	srcIsDir      bool
}

// implement copy objects
func (b *BosCli) copyBetweenRemote(srcPath, dstPath, storageClass string, recursive,
	restart bool) (BosCliErrorCode, error) {
	// preprocessing and check request
	args, retCode, err := b.copyRemoteRequestPreProcess(srcPath, dstPath, storageClass, recursive)
	if err != nil {
		return retCode, err
	}

	// execute copy between remote
	ret, retCode, err := b.copyObjectExecute(args, storageClass, restart)

	// print result
	if err != nil {
		bcecliAbnormalExistCodeErr(retCode, err)
	}
	// TODO need print failed number
	printIfNotQuiet("[%d] objects remote copied.\n", ret.successed)
	if ret.failed > 0 {
		return BOSCLI_EMPTY_CODE, nil
	} else {
		return BOSCLI_OK, nil
	}
}

// check whether the request of copy between is valid, and get request info
func (b *BosCli) copyRemoteRequestPreProcess(srcPath, dstPath, storageClass string,
	recursive bool) (*copyBetweenRemoteArgs, BosCliErrorCode, error) {

	srcBucketName, srcObjectKey := splitBosBucketKey(srcPath)
	dstBucketName, dstObjectKey := splitBosBucketKey(dstPath)

	_, retCode := getStorageClassFromStr(storageClass)
	if retCode != BOSCLI_OK {
		return nil, retCode, fmt.Errorf("don't support storage-class %s", storageClass)
	}

	if srcBucketName == "" {
		return nil, BOSCLI_SRC_BUCKET_IS_EMPTY, fmt.Errorf("Please check source bucket name")
	}
	if dstBucketName == "" {
		return nil, BOSCLI_DST_BUCKET_IS_EMPTY, fmt.Errorf("Please check destination bucket name")
	}

	ok, err := b.handler.doesBucketExist(b.bosClient, srcBucketName)
	if err != nil {
		return nil, BOSCLI_EMPTY_CODE, err
	} else if !ok {
		return nil, BOSCLI_SRC_BUCKET_DONT_EXIST, fmt.Errorf("source bucket %s don't exist!",
			srcBucketName)
	}

	ok, err = b.handler.doesBucketExist(b.bosClient, dstBucketName)
	if err != nil {
		return nil, BOSCLI_EMPTY_CODE, err
	} else if !ok {
		return nil, BOSCLI_DST_BUCKET_DONT_EXIST, fmt.Errorf("destination bucket %s don't exist!",
			srcBucketName)
	}

	// src is dir
	if srcObjectKey == "" || strings.HasSuffix(srcObjectKey, boscmd.BOS_PATH_SEPARATOR) {
		if dstObjectKey != "" && !strings.HasSuffix(dstObjectKey, boscmd.BOS_PATH_SEPARATOR) {
			return nil, BOSCLI_BATCH_COPY_DSTOBJECT_END, fmt.Errorf("Can not copy '%s/%s' to an"+
				"object '%s'", srcBucketName, srcObjectKey, dstObjectKey)
		}
		if recursive {
			return &copyBetweenRemoteArgs{
				srcBucketName: srcBucketName,
				srcObjectKey:  srcObjectKey,
				dstBucketName: dstBucketName,
				dstObjectKey:  dstObjectKey,
				srcIsDir:      true,
			}, BOSCLI_OK, nil
		} else {
			return nil, BOSCLI_BATCH_COPY_SRCOBJECT_END, fmt.Errorf("Please use -r to copy " +
				"folder between bos")
		}
	}

	// src is single object
	return &copyBetweenRemoteArgs{
		srcBucketName: srcBucketName,
		srcObjectKey:  srcObjectKey,
		dstBucketName: dstBucketName,
		dstObjectKey:  dstObjectKey,
	}, BOSCLI_OK, nil
}

// copy objects
func (b *BosCli) copyObjectExecute(args *copyBetweenRemoteArgs, storageClass string, restart bool) (
	*executeResult, BosCliErrorCode, error) {

	var (
		copied        int
		failedNum     int
		listResult    *listFileResult
		srcObjectName string
		dstObjectName string
		srcBosClient  bosClientInterface
		err           error
	)

	// generate object list iterator
	if srcBosClient, err = initBosClientForBucket("", "", args.srcBucketName); err != nil {
		return nil, BOSCLI_EMPTY_CODE, err
	}
	objectLists := NewObjectListIterator(srcBosClient, nil, args.srcBucketName, args.srcObjectKey,
		"", true, true, args.srcIsDir, false, 1000)

	for {
		listResult, err = objectLists.next()
		if err != nil {
			return &executeResult{successed: copied, failed: failedNum}, BOSCLI_EMPTY_CODE, err
		}
		if listResult.ended {
			break
		}
		if listResult.isDir {
			continue
		}

		object := listResult.file
		srcObjectName = object.path
		dstObjectName = object.key
		if dstObjectName == "" {
			continue
		}

		if args.dstObjectKey != "" {
			if strings.HasSuffix(args.dstObjectKey, boscmd.BOS_PATH_SEPARATOR) {
				dstObjectName = args.dstObjectKey + dstObjectName
			} else {
				dstObjectName = args.dstObjectKey
			}
		}

		if isTheSameBucketAndObject(args.srcBucketName, srcObjectName, args.dstBucketName,
			dstObjectName, storageClass, object.storageClass) {
			failedNum++
			printIfNotQuiet("Can not cover object with same object, skip: %s\n", object.key)
			continue
		}
		err = b.handler.utilCopyObject(srcBosClient, b.bosClient, args.srcBucketName, srcObjectName,
			args.dstBucketName, dstObjectName, storageClass, object.size, object.mtime,
			object.gtime, restart)
		if err == nil {
			copied++
		} else {
			failedNum++
			fmt.Printf("Error occurs when copy object %s%s/%s: %s\n", BOS_PATH_PREFIX,
				args.srcBucketName, srcObjectName, getErrorMsg(err))
		}
	}
	return &executeResult{successed: copied, failed: failedNum}, BOSCLI_EMPTY_CODE, err
}

type copyDownloadArgs struct {
	srcBucketName      string
	srcObjectKey       string
	srcIsDir           bool
	isDownloadToStream bool
}

// implement downlaod object
func (b *BosCli) copyDownload(srcPath, dstPath, downLoadTmp string, recursive, yes,
	restart bool) (BosCliErrorCode, error) {
	// preprocessing request
	args, retCode, err := b.copyDownloadPreProcess(srcPath, dstPath, recursive)
	if err != nil {
		return retCode, err
	}

	if args.isDownloadToStream {
		return BOSCLI_EMPTY_CODE, fmt.Errorf("download to stream is not implement")
	}

	// generate oplist and execute download
	ret, retCode, err := b.copyDownloadExecute(args, dstPath, downLoadTmp, yes, restart)

	// print result
	if err != nil {
		return retCode, err
	}
	// TODO need print failed number
	printIfNotQuiet("[%d] objects downloaded.\n", ret.successed)
	if ret.failed > 0 {
		return BOSCLI_EMPTY_CODE, nil
	} else {
		return BOSCLI_OK, nil
	}
}

// preprocessing from download files
// RETURN:
//    srcIsDir: whether srcPath is directory or file
//    error code and error
func (b *BosCli) copyDownloadPreProcess(srcPath, dstPath string, recursive bool) (*copyDownloadArgs,
	BosCliErrorCode, error) {

	srcBucketName, srcObjectKey := splitBosBucketKey(srcPath)
	if srcBucketName == "" {
		return nil, BOSCLI_SRC_BUCKET_IS_EMPTY, fmt.Errorf("Please check source bucket name")
	}
	ok, err := b.handler.doesBucketExist(b.bosClient, srcBucketName)
	if err != nil {
		return nil, BOSCLI_EMPTY_CODE, err
	} else if !ok {
		return nil, BOSCLI_SRC_BUCKET_DONT_EXIST, fmt.Errorf("bucekt %s don't exist!",
			srcBucketName)
	}

	if util.DoesDirExist(dstPath) && !util.IsDirWritable(dstPath) {
		return nil, BOSCLI_DIR_IS_NOT_WRITABLE, fmt.Errorf("Directory %s is not writable!",
			dstPath)
	}

	// check batch download
	if srcObjectKey == "" || strings.HasSuffix(srcObjectKey, boscmd.BOS_PATH_SEPARATOR) {
		if util.DoesFileExist(dstPath) {
			return nil, BOSCLI_CANT_DOWNLOAD_FILES_TO_FILE, fmt.Errorf("Can not download files"+
				"to a file %s", dstPath)
		}
		if recursive {
			return &copyDownloadArgs{
				srcBucketName: srcBucketName,
				srcObjectKey:  srcObjectKey,
				srcIsDir:      true,
			}, BOSCLI_OK, nil
		} else {
			return nil, BOSCLI_BATCH_DOWNLOAD_SRCOBJECT_END, fmt.Errorf("Please use -r to " +
				"download objects")
		}
	}

	// cehck downlaod to stream
	if dstPath == "-" {
		return nil, BOSCLI_UNSUPPORT_METHOD, fmt.Errorf("CLI don't yet support Download to stream")
	}

	return &copyDownloadArgs{
		srcBucketName: srcBucketName,
		srcObjectKey:  srcObjectKey,
	}, BOSCLI_OK, nil
}

// generate oplist and execute download
func (b *BosCli) copyDownloadExecute(args *copyDownloadArgs, dstPath, downLoadTmp string, yes,
	restart bool) (*executeResult, BosCliErrorCode, error) {

	var (
		downloaded int
		failedNum  int
		listResult *listFileResult
		retCode    BosCliErrorCode = BOSCLI_EMPTY_CODE
		err        error
	)

	// download to stream
	if dstPath == "-" {
		return nil, retCode, fmt.Errorf("unsupport upload from stream!")
	}

	// download single object
	if !args.srcIsDir {
		err = b.handler.utilDownloadObject(b.bosClient, args.srcBucketName, args.srcObjectKey,
			dstPath, downLoadTmp, yes, 0, 0, 0, restart)

		if err == nil {
			retCode = BOSCLI_OK
			downloaded++
		}
		return &executeResult{successed: downloaded, failed: failedNum}, retCode, err
	}

	// batch download
	objectList := NewObjectListIterator(b.bosClient, nil, args.srcBucketName, args.srcObjectKey,
		"", true, true, true, false, 1000)
	for {
		listResult, err = objectList.next()
		if err != nil {
			break
		}
		if listResult.ended {
			break
		}
		if listResult.isDir {
			continue
		}

		object := listResult.file
		srcObjectName := object.path
		tmpDstFileName := object.key
		if tmpDstFileName == "" {
			continue
		}
		if strings.HasSuffix(srcObjectName, boscmd.BOS_PATH_SEPARATOR) {
			continue
		}

		dstFileName := dstPath
		if strings.HasSuffix(dstPath, util.OsPathSeparator) {
			dstFileName += tmpDstFileName
		} else {
			dstFileName += util.OsPathSeparator + tmpDstFileName
		}

		err = b.handler.utilDownloadObject(b.bosClient, args.srcBucketName, srcObjectName,
			dstFileName, downLoadTmp, yes, object.size, object.mtime, object.gtime, restart)

		if err == nil {
			downloaded++
		} else {
			failedNum++
			fmt.Printf("Error occurs when download object %s%s/%s: %s\n", BOS_PATH_PREFIX,
				args.srcBucketName, srcObjectName, getErrorMsg(err))
		}
	}
	return &executeResult{successed: downloaded, failed: failedNum}, retCode, err
}

type copyUploadArges struct {
	srcPath          string
	dstBucketName    string
	dstObjectKey     string
	srcIsDir         bool
	uploadFromStream bool
}

func (b *BosCli) copyUpload(srcPath, dstPath, storageClass string, recursive,
	restart bool) (BosCliErrorCode, error) {
	// preprocessing and check request
	args, retCode, err := b.copyUploadRequestPreProcess(srcPath, dstPath, storageClass, recursive)
	if err != nil {
		return retCode, err
	}

	if args.uploadFromStream {
		return BOSCLI_EMPTY_CODE, fmt.Errorf("upload from stream is not implement")
	}

	// execute upload file to bos
	ret, retCode, err := b.uploadFileExecute(args, srcPath, storageClass, restart)

	// print result
	if err != nil {
		return retCode, err
	}
	// TODO need print failed number
	printIfNotQuiet("[%d] objects uploaded.\n", ret.successed)
	if ret.failed > 0 {
		return BOSCLI_EMPTY_CODE, nil
	} else {
		return BOSCLI_OK, nil
	}
}

// preprocessing and check upload request
func (b *BosCli) copyUploadRequestPreProcess(srcPath, dstPath, storageClass string,
	recursive bool) (*copyUploadArges, BosCliErrorCode, error) {

	dstBucketName, dstObjectKey := splitBosBucketKey(dstPath)
	if dstBucketName == "" {
		return nil, BOSCLI_DST_BUCKET_IS_EMPTY, fmt.Errorf("Bucket name should not be empty ")
	}

	// verify storage class
	storageClass, retCode := getStorageClassFromStr(storageClass)
	if retCode != BOSCLI_OK {
		return nil, retCode, fmt.Errorf("don't support storage-class %s", storageClass)
	}

	// single file
	if srcPath == "-" {
		if dstObjectKey == "" {
			return nil, BOSCLI_DST_OBJECT_KEY_IS_EMPTY, fmt.Errorf("Object key can not be empty!")
		}
		if strings.HasSuffix(dstObjectKey, boscmd.BOS_PATH_SEPARATOR) {
			return nil, BOSCLI_UPLOAD_STREAM_TO_DIR, fmt.Errorf("Can not upload stream to path")
		}
		return nil, BOSCLI_UNSUPPORT_METHOD, fmt.Errorf("CLI don't yet support upload from stream")
	}

	if !util.DoesPathExist(srcPath) {
		return nil, boscmd.LOCAL_PATH_NOT_EXIST, fmt.Errorf("Source path %s does not exist!",
			srcPath)
	}

	// check whether the dst bucekt exist
	ok, err := b.handler.doesBucketExist(b.bosClient, dstBucketName)
	if err != nil {
		return nil, BOSCLI_EMPTY_CODE, err
	} else if !ok {
		return nil, BOSCLI_DST_BUCKET_DONT_EXIST, fmt.Errorf("bucekt %s don't exist!",
			dstBucketName)
	}

	// batch files
	if recursive {
		if !util.DoesFileExist(srcPath) {
			if dstObjectKey != "" && !strings.HasSuffix(dstObjectKey, boscmd.BOS_PATH_SEPARATOR) {
				dstObjectKey += boscmd.BOS_PATH_SEPARATOR
			}
			return &copyUploadArges{
				srcPath:       srcPath,
				dstBucketName: dstBucketName,
				dstObjectKey:  dstObjectKey,
				srcIsDir:      true,
			}, BOSCLI_OK, nil
		}
	}

	// single file
	if util.DoesDirExist(srcPath) {
		return nil, BOSCLI_UPLOAD_SRC_CANNT_BE_DIR, fmt.Errorf("Cannot upload directory %s.",
			srcPath)
	}
	return &copyUploadArges{
		srcPath:       srcPath,
		dstBucketName: dstBucketName,
		dstObjectKey:  dstObjectKey,
	}, BOSCLI_OK, nil
}

// generate oplist and excute upload
func (b *BosCli) uploadFileExecute(args *copyUploadArges, srcPath string, storageClass string,
	restart bool) (*executeResult, BosCliErrorCode, error) {

	var (
		listResult *listFileResult
		uploaded   int
		failedNum  int
		err        error
	)

	// upload from stream
	if srcPath == "-" {
		return nil, BOSCLI_UNSUPPORT_METHOD, fmt.Errorf("CLI don't yet support upload from stream")
	}

	// generate object list iterator
	absSrcPath, _ := util.Abs(srcPath)
	filesList := NewLocalFileIterator(absSrcPath, nil, true)

	// upload from file
	for {
		listResult, err = filesList.next()
		if err != nil {
			break
		}
		if listResult.err != nil {
			err = listResult.err
			break
		}
		if listResult.ended {
			break
		}

		file := listResult.file

		if file.err != nil {
			failedNum++
			printIfNotQuiet("Failed Upload: %s. Receive error: %s\n", file.path, file.err.Error())
			continue
		}

		// get final object key
		finalObjectKey := getFinalObjectKeyFromLocalPath(absSrcPath, file.path, args.dstObjectKey,
			args.srcIsDir)

		//excute upload
		err = b.handler.utilUploadFile(b.bosClient, file.path, file.realPath, args.dstBucketName,
			finalObjectKey, storageClass, file.size, file.mtime, file.gtime, restart)

		if err != nil {
			printIfNotQuiet("Failed Upload: %s to %s%s/%s. Receive error: %s\n", file.path,
				BOS_PATH_PREFIX, args.dstBucketName, finalObjectKey, err.Error())
			failedNum++
		} else {
			uploaded++
		}
	}
	return &executeResult{successed: uploaded, failed: failedNum}, BOSCLI_EMPTY_CODE, err
}

type syncArgs struct {
	srcPath              string
	dstPath              string
	srcBucketName        string
	srcObjectKey         string
	dstBucketName        string
	dstObjectKey         string
	srcType              string
	dstType              string
	syncType             string
	syncProcessingNum    int
	multiUploadThreadNum int64
}

// sync local folder to bos
// 1. list all src
// 2. list all dst
// 3. compare and gen file list of src to be put to dst, and src to delete, if delete is defined
// 4. if dryrun is defined, show list to be processed
// param args: parsed args, must have SRC and DST explicitly defined
func (b *BosCli) Sync(srcPath, dstPath, storageClass, downLoadTmp, syncType string, exclude, include, excludeTime,
	includeTime, excludeDelete []string, concurrency int, del, dryrun, yes, quiet, disableBar, restart bool) {

	var (
		filter       *bosFilter = nil
		deleteFilter *bosFilter = nil
		retCode      BosCliErrorCode
		err          error
	)
	Quiet = quiet
	DisableBar = disableBar
	IsConcurrentOperation = true

	// preprocessing for sync reques
	args, retCode, err := b.syncPreProcess(srcPath, dstPath, storageClass, exclude, include,
		excludeTime, includeTime, concurrency, del, yes)
	if err != nil {
		bcecliAbnormalExistCodeErr(retCode, err)
	}

	//generate new filter
	if len(exclude) > 0 || len(include) > 0 || len(excludeTime) > 0 || len(includeTime) > 0 {
		filter, retCode, err = newSyncFilter(exclude, include, excludeTime, includeTime,
			args.srcType == IS_LOCAL)
		if err != nil {
			bcecliAbnormalExistCodeErr(retCode, err)
		}
	}
	if del && len(excludeDelete) > 0 {
		deleteFilter, retCode, err = newSyncFilter(excludeDelete, []string{}, []string{},
			[]string{}, args.dstType == IS_LOCAL)
	}

	result, retCode, err := b.syncExecute(filter, deleteFilter, args, storageClass, downLoadTmp,
		syncType, del, dryrun, restart)
	if err != nil {
		if result != nil {
			printIfNotQuiet("Sync interrupted: %s to %s, [%d] success, [%d] failure\n",
				args.srcPath, args.dstPath, result.successed, result.failed)
		}
		bcecliAbnormalExistCodeErr(retCode, err)
	}
	printIfNotQuiet("Sync done: %s to %s, [%d] success, [%d] failure\n", args.srcPath, args.dstPath,
		result.successed, result.failed)
	if result.failed > 0 {
		bcecliAbnormalExistCode(BOSCLI_EMPTY_CODE)
	}
}

func (b *BosCli) syncPreProcess(srcPath, dstPath, storageClass string, exclude, include,
	excludeTime, includeTime []string, concurrency int, del, yes bool) (*syncArgs,
	BosCliErrorCode, error) {

	var (
		ok bool
	)

	// exclude and include time cannot be used together
	if len(exclude) != 0 && len(include) != 0 {
		return nil, BOSCLI_SYNC_EXCLUDE_INCLUDE_TOG, fmt.Errorf("exclude and include" +
			"time cannot be used together")
	}
	// exclude time and include time cannot be used together
	if len(excludeTime) != 0 && len(includeTime) != 0 {
		return nil, BOSCLI_SYNC_EXCLUDE_INCLUDE_TIME_TOG, fmt.Errorf("exclude time and include" +
			"time cannot be used together")
	}

	args := &syncArgs{}

	// check src path
	if strings.HasPrefix(srcPath, BOS_PATH_PREFIX) {
		args.srcType = IS_BOS
		args.srcPath = trimTrailingSlash(srcPath) + boscmd.BOS_PATH_SEPARATOR
		args.srcBucketName, args.srcObjectKey = splitBosBucketKey(args.srcPath)
	} else {
		args.srcType = IS_LOCAL
		args.srcPath = trimTrailingSlash(srcPath) + util.OsPathSeparator
		if !util.DoesPathExist(srcPath) {
			return nil, boscmd.LOCAL_PATH_NOT_EXIST, fmt.Errorf("file path %s doesn't exist",
				srcPath)
		} else if !util.DoesDirExist(srcPath) {
			return nil, BOSCLI_SYNC_UPLOAD_SRC_MUST_DIR, fmt.Errorf("SRC must be a local folder.")
		}
	}

	// check dst path
	if strings.HasPrefix(dstPath, BOS_PATH_PREFIX) {
		args.dstType = IS_BOS
		args.dstPath = trimTrailingSlash(dstPath) + boscmd.BOS_PATH_SEPARATOR
		args.dstBucketName, args.dstObjectKey = splitBosBucketKey(args.dstPath)
	} else {
		args.dstType = IS_LOCAL
		args.dstPath = trimTrailingSlash(dstPath) + util.OsPathSeparator
		if util.DoesPathExist(dstPath) && !util.DoesDirExist(dstPath) {
			return nil, BOSCLI_SYNC_DOWN_DST_MUST_DIR, fmt.Errorf("DST must be a local folder.")
		}
	}

	args.syncType = args.srcType + args.dstType
	if args.syncType == LOCAL_TO_LOCAL {
		return nil, BOSCLI_SYNC_LOCAL_TO_LOCAL, fmt.Errorf("can't sync local -> local")
	}

	// get concurrency of sync
	if concurrency < 0 {
		return nil, BOSCLI_SYNC_PROCESS_NUM_LESS_ZERO, fmt.Errorf("concurrency for sync must" +
			"greater than zero")
	} else if concurrency == 0 {
		concurrency, ok = bceconf.ServerConfigProvider.GetSyncProcessingNum()
		if !ok {
			return nil, BOSCLI_GET_SYNC_PROCESSING_NUM_FAILED, fmt.Errorf("There is no info " +
				"about sync processing num found!")
		}
	}
	args.syncProcessingNum = concurrency

	// get multi upload or copy thread num
	args.multiUploadThreadNum, ok = bceconf.ServerConfigProvider.GetMultiUploadThreadNum()
	if !ok {
		return nil, BOSCLI_GET_UPLOAD_THREAD_NUM_FAILED, fmt.Errorf("There is no info about " +
			"multi upload thread Num found!")
	}

	// when filter and delete are used together, bcecmd will delete the filtered object in
	// destination, need user to confirm
	if (len(exclude) != 0 || len(include) != 0 || len(excludeTime) != 0 || len(includeTime) != 0) &&
		del && !yes {
		yes = util.PromptConfirm("NOTICE: when filtering policy and delete are used together, " +
			"BCECMD will delete the filtered objects on destination, but them may exist on " +
			"source, Do you really want to REMOVE the filtered objects?")
		if !yes {
			return nil, BOSCLI_OPRATION_CANCEL, fmt.Errorf("")
		}
	} else if del && !yes {
		// confirm delete
		yes = util.PromptConfirm("Do you really want to REMOVE objects do not exist on %s but "+
			"exist on %s?", srcPath, dstPath)
		if !yes {
			return nil, BOSCLI_OPRATION_CANCEL, fmt.Errorf("")
		}
	}

	return args, BOSCLI_OK, nil
}

func (b *BosCli) syncExecute(filter, deleteFilter *bosFilter, args *syncArgs, storageClass, downLoadTmp,
	syncType string, del, dryrun, restart bool) (*executeResult, BosCliErrorCode, error) {

	var (
		srcBosClient bosClientInterface
		srcFiles     fileListIterator
		dstFiles     fileListIterator
		atBothSide   syncStrategyInfterface
		notAtSrc     syncStrategyInfterface
		opSync       sync.WaitGroup
		retErr       error
		err          error
	)

	// get src file list iterator
	if args.srcType == IS_LOCAL {
		if absSrcPath, err := util.Abs(args.srcPath); err != nil {
			return nil, BOSCLI_EMPTY_CODE, err
		} else {
			srcFiles = NewLocalFileIterator(absSrcPath, filter, true)
		}
	} else if args.srcType == IS_BOS {
		if srcBosClient, err = initBosClientForBucket("", "", args.srcBucketName); err != nil {
			return nil, BOSCLI_EMPTY_CODE, err
		} else {
			srcFiles = NewObjectListIterator(srcBosClient, filter, args.srcBucketName,
				args.srcObjectKey, "", true, true, true, false, 1000)
		}
	} else {
		return nil, BOSCLI_EMPTY_CODE, fmt.Errorf("Unknown source type!")
	}

	// get dst file list iterator
	if args.dstType == IS_LOCAL {
		if absDstPath, err := util.Abs(args.dstPath); err != nil {
			return nil, BOSCLI_EMPTY_CODE, err
		} else {
			dstFiles = NewLocalFileIterator(absDstPath, nil, true)
		}
	} else if args.dstType == IS_BOS {
		dstFiles = NewObjectListIterator(b.bosClient, nil, args.dstBucketName, args.dstObjectKey,
			"", true, true, true, false, 1000)
	} else {
		return nil, BOSCLI_EMPTY_CODE, fmt.Errorf("Unknown destination type!")
	}

	// init sync strategies
	if syncType == "" || syncType == "time-size" {
		atBothSide = &sizeAndLastModifiedSync{}
	} else if syncType == "time-size-crc32" {
		atBothSide = &sizeAndLastModifiedAndCrc32Sync{
			srcType:       args.srcType,
			dstType:       args.dstType,
			srcBucketName: args.srcBucketName,
			dstBucketName: args.dstBucketName,
			srcBosClient:  srcBosClient,
			dstBosClient:  b.bosClient,
		}
	} else if syncType == "only-crc32" {
		atBothSide = &crc32Sync{
			srcType:       args.srcType,
			dstType:       args.dstType,
			srcBucketName: args.srcBucketName,
			dstBucketName: args.dstBucketName,
			srcBosClient:  srcBosClient,
			dstBosClient:  b.bosClient,
		}
	} else {
		return nil, BOSCLI_INVALID_SYNY_TYPE, fmt.Errorf("Unknown sync type!")
	}

	notAtDst := &alwaysSync{}
	if del {
		notAtSrc = &deleteDstSync{
			deleteFilter:  deleteFilter,
			dstType:       args.dstType,
			dstBucketName: args.dstBucketName,
		}
	}
	comparator := NewComparator(atBothSide, notAtDst, notAtSrc, args, srcFiles, dstFiles)

	// init channel
	syncOpPool := make(chan int, args.syncProcessingNum)
	executeResultChan := make(chan int, args.syncProcessingNum)
	isEnd := make(chan bool, 1)
	syncResultChan := make(chan executeResult, 1)
	overWriteDst := true

	// define a function to count the failed and successed number of sync operation
	syncResultCount := func(syncResultChan chan executeResult) {
		failed := 0
		successed := 0
		for {
			select {
			case result := <-executeResultChan:
				if result == -1 { // fail
					failed += 1
				} else if result == 1 { // success
					successed += 1
				}
			case <-isEnd:
				syncResultChan <- executeResult{failed: failed, successed: successed}
				return
			}
		}
	}

	// this function is used to execute sync operation
	syncOpFunc := func(syncInfo *syncOpDetail, flag, prompt string, wg *sync.WaitGroup) {
		var (
			err error
		)

		defer func() {
			wg.Done()
			<-syncOpPool
		}()

		switch flag {
		case SYNC_OP_COPY:
			err = b.handler.utilCopyObject(srcBosClient, b.bosClient, args.srcBucketName,
				syncInfo.srcPath, args.dstBucketName, syncInfo.dstPath, storageClass,
				syncInfo.srcFileInfo.size, syncInfo.srcFileInfo.mtime, syncInfo.srcFileInfo.gtime,
				restart)

		case SYNC_OP_UPLOAD:
			err = b.handler.utilUploadFile(b.bosClient, syncInfo.srcPath,
				syncInfo.srcFileInfo.realPath, args.dstBucketName, syncInfo.dstPath, storageClass,
				syncInfo.srcFileInfo.size, syncInfo.srcFileInfo.mtime, syncInfo.srcFileInfo.gtime,
				restart)

		case SYNC_OP_DOWNLOAD:
			err = b.handler.utilDownloadObject(b.bosClient, args.srcBucketName, syncInfo.srcPath,
				syncInfo.dstPath, downLoadTmp, overWriteDst, syncInfo.srcFileInfo.size,
				syncInfo.srcFileInfo.mtime, syncInfo.srcFileInfo.gtime, restart)

		case SYNC_OP_REMOVE:
			err = b.handler.utilDeleteObject(b.bosClient, args.dstBucketName, syncInfo.dstPath)

		case SYNC_OP_DELETE:
			err = b.handler.utilDeleteLocalFile(syncInfo.dstPath)

		default:
			err = fmt.Errorf("Sync destination is a folder instead of a file: %s", args.dstPath)
		}

		// send result to syncResultCount
		if err != nil {
			executeResultChan <- -1 // -1 represent failed
			printIfNotQuiet("Failed %s. Error: %s\n", prompt, err)
		} else {
			executeResultChan <- 1 // 1 represent successed
		}
	}

	// start syncResultCount
	go syncResultCount(syncResultChan)

	for {
		syncInfo, err := comparator.next()
		if err != nil {
			retErr = err
			break
		}
		if syncInfo.ended {
			break
		}
		if syncInfo.err != nil {
			if ErrIsNotExist(syncInfo.err) {
				if !dryrun {
					executeResultChan <- -1
				}
				printIfNotQuiet("Failed: %s. It may have been deleted!\n", syncInfo.err)
				continue
			} else {
				retErr = syncInfo.err
				break
			}
		}

		if args.syncType == BOS_TO_LOCAL && strings.HasSuffix(syncInfo.srcPath,
			boscmd.BOS_PATH_SEPARATOR) {
			continue
		}

		prompt := ""
		flag := ""
		switch syncInfo.syncFunc {
		// copy, upload or download
		case OPERATE_CMD_COPY:
			if args.syncType == LOCAL_TO_BOS {
				flag = SYNC_OP_UPLOAD
				prompt = fmt.Sprintf("%s: %s to bos:/%s/%s", flag, syncInfo.srcPath,
					args.dstBucketName, syncInfo.dstPath)
			} else if args.syncType == BOS_TO_LOCAL {
				flag = SYNC_OP_DOWNLOAD
				prompt = fmt.Sprintf("%s: bos:/%s/%s to %s", flag, args.srcBucketName,
					syncInfo.srcPath, syncInfo.dstPath)
			} else if args.syncType == BOS_TO_BOS {
				flag = SYNC_OP_COPY
				prompt = fmt.Sprintf("%s: bos:/%s/%s to bos:/%s/%s", flag, args.srcBucketName,
					syncInfo.srcPath, args.dstBucketName, syncInfo.dstPath)
			} else {
				retErr = fmt.Errorf("Unknown sync operation!")
				goto END
			}
			if args.dstType == IS_LOCAL && util.DoesPathExist(syncInfo.dstPath) &&
				util.DoesDirExist(syncInfo.dstPath) {
				flag = SYNC_OP_ERROR
			}
		// delete local file or bos object
		case OPERATE_CMD_DELETE:
			if args.dstType == IS_BOS {
				flag = SYNC_OP_REMOVE
				prompt = fmt.Sprintf("%s: bos:/%s/%s", flag, args.dstBucketName, syncInfo.dstPath)
			} else if args.dstType == IS_LOCAL {
				flag = SYNC_OP_DELETE
				prompt = fmt.Sprintf("%s: %s", flag, syncInfo.dstPath)
			}
		default:
			continue
		}

		if dryrun && prompt != "" {
			printIfNotQuiet("%s\n", prompt)
		} else {
			syncOpPool <- 1
			opSync.Add(1)
			go syncOpFunc(syncInfo, flag, prompt, &opSync)
		}
	}

END:
	// waiting for all sync operation finish
	opSync.Wait()

	// end count result
	// Sleep 200ms, waiting for syncResultCount add all op results
	time.Sleep(200 * time.Millisecond)
	isEnd <- true
	ret := <-syncResultChan
	return &ret, BOSCLI_EMPTY_CODE, retErr
}
