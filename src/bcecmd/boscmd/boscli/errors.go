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

// Codes in this file is used to process bos cli error massage.

package boscli

import (
	"fmt"
	"os"
)

import (
	"bcecmd/boscmd"
	"github.com/baidubce/bce-sdk-go/bce"
)

type BosCliErrorCode string

const (
	BOSCLI_SUGGETION_PROPMT = "提示"
)

const (
	BOSCLI_OK                                 = "boscliSuccess"
	BOSCLI_EMPTY_CODE                         = ""
	BOSCLI_OPRATION_CANCEL                    = "boscliOperationCancel"
	BOSCLI_UNSUPPORT_METHOD                   = "boscliUnsupportMethod"
	BOSCLI_BUCKETNAME_IS_EMPTY                = "boscliBucketNameIsEmpty"
	BOSCLI_OBJECTKEY_IS_EMPTY                 = "boscliObjectKeyIsEmpty"
	BOSCLI_BOSPATH_IS_INVALID                 = "boscliBosPathIsInvalid"
	BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME      = "boscliBucketNameContainObjectName"
	BOSCLI_BUCKET_NOT_EMPTY                   = "BucketNotEmpty"
	BOSCLI_INTERNAL_ERROR                     = "boscliInternalError"
	BOSCLI_UNSUPPORT_STORAGE_CLASS            = "boscliUnsupportStorageClass"
	BOSCLI_SRC_BUCKET_IS_EMPTY                = "boscliSrcBucketIsEmpty"
	BOSCLI_DST_BUCKET_IS_EMPTY                = "boscliDstBucketIsEmpty"
	BOSCLI_SRC_BUCKET_DONT_EXIST              = "boscliSrcBucketNotExist"
	BOSCLI_DST_BUCKET_DONT_EXIST              = "boscliSrcBucketNotExist"
	BOSCLI_SRC_OBJECT_IS_EMPTY                = "boscliSrcObjectIsEmpty"
	BOSCLI_BATCH_COPY_DSTOBJECT_END           = "boscliBathcCopyDstObjectEnd"
	BOSCLI_BATCH_COPY_SRCOBJECT_END           = "boscliBathcCopySrcObjectEnd"
	BOSCLI_BATCH_UPLOAD_SRC_PATH_END          = "boscliBatchUploadSrcPathEnd"
	BOSCLI_COVER_SELF                         = "boscliCopyCoverSelf"
	BOSCLI_CANT_DOWNLOAD_FILES_TO_FILE        = "boscliCantDownloadFilesToFile"
	BOSCLI_DIR_IS_NOT_WRITABLE                = "boscliDirIsNotWritable"
	BOSCLI_BATCH_DOWNLOAD_SRCOBJECT_END       = "boscliBatchDownlaodSrcObjectEnd"
	BOSCLI_UPLOAD_SRC_CANNT_BE_DIR            = "boscliUploadSrcCanntBeDir"
	BOSCLI_DST_OBJECT_KEY_IS_EMPTY            = "boscliDstObjectKeyIsEmpty"
	BOSCLI_UPLOAD_STREAM_TO_DIR               = "boscliUploadStreamToDir"
	BOSCLI_RM_DIR_MUST_USE_RECURSIVE          = "boscliRmDirMustUseRecursive"
	BOSCLI_EXPIRE_LESS_NONE                   = "boscliExpireLessNegativeOne"
	BOSCLI_SYNC_EXCLUDE_INCLUDE_TIME_TOG      = "boscliSyncExcludeIncludeTimeTog"
	BOSCLI_SYNC_EXCLUDE_INCLUDE_TOG           = "boscliSyncEcludeIncludeTog"
	BOSCLI_SYNC_UPLOAD_SRC_MUST_DIR           = "boscliSyncUploadSrcMustDir"
	BOSCLI_SYNC_DOWN_DST_MUST_DIR             = "boscliSyncDownDstMustDir"
	BOSCLI_SYNC_LOCAL_TO_LOCAL                = "boscliSyncLocalToLocal"
	BOSCLI_SYNC_PROCESS_NUM_LESS_ZERO         = "boscliSyncProcessNumLessZero"
	BOSCLI_INVALID_SYNY_TYPE                  = "boscliInvalidSyncType"
	BOSCLI_GET_SYNC_PROCESSING_NUM_FAILED     = "boscliGetUploadProcessingNumFailed"
	BOSCLI_GET_UPLOAD_THREAD_NUM_FAILED       = "boscliGetUplaodThreadNumFailed"
	BOSCLI_PUT_LIFECYCLE_NO_CONFIG_AND_BUCKET = "boscliPutLifecycleNoConfigAndBucket"
	BOSCLI_PUT_LOG_NO_TARGET_BUCKET           = "boscliPutLogNoTargetBucket"
	BOSCLI_STORAGE_CLASS_IS_EMPTY             = "boscliStorageClasssIsEmpty"
	BOSCLI_PUT_ACL_CANNED_FILE_SAME_TIME      = "boscliPutAclCannedFileSameTime"
	BOSCLI_PUT_ACL_CANNED_FILE_BOTH_EMPTY     = "boscliPutAclCannedFileBothEmpty"
	BOSCLI_PUT_ACL_CANNED_DONT_SUPPORT        = "boscliPutAclCannedDontSupport"
)

var BosCliSuggetions map[BosCliErrorCode]string

func init() {
	initSuggetionChinese()
}

func initSuggetionChinese() {
	BosCliSuggetions = make(map[BosCliErrorCode]string)
	BosCliSuggetions[BOSCLI_UNSUPPORT_METHOD] =
		"BOS CLI 还不支持此方法！"
	BosCliSuggetions[BOSCLI_BUCKETNAME_IS_EMPTY] =
		"您输入的 Bucket Name 不能为空！"
	BosCliSuggetions[BOSCLI_OBJECTKEY_IS_EMPTY] =
		"您输入的 Object Key 不能为空！"
	BosCliSuggetions[BOSCLI_BOSPATH_IS_INVALID] =
		"请您检查是否输入了无效的Bos路径！BOS路径可以是bucket 也可以是 bucket + object, Bos路径" +
			"必须以 bos:/开始。\n" +
			"例如: bos:/bucket, bos:/bucket/object。"
	BosCliSuggetions[BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME] =
		"Bucket Name 不应该包含 Object 信息（Bucket Name 只能包含小写字母、数字和“-”，开头结" +
			"尾为小写字母和数字，长度在3-63之间）"
	BosCliSuggetions[BOSCLI_BUCKET_NOT_EMPTY] =
		"你要删除的Bucket 不为空!, 您可以加上参数 -f 强制删除bucket.\n" +
			"例如： bcecmd bos rb bos:/bucket -f"
	BosCliSuggetions[BOSCLI_INTERNAL_ERROR] =
		"CLI 内部发生了错误，请重试！"
	BosCliSuggetions[BOSCLI_UNSUPPORT_STORAGE_CLASS] =
		"BOS 当前只支持 STANDARD（标准）， STANDARD_IA（低频）和 COLD（冷存储） 三种存储类型！"
	BosCliSuggetions[BOSCLI_SRC_BUCKET_IS_EMPTY] =
		"请检查源 bucket 名字是否为空！"
	BosCliSuggetions[BOSCLI_DST_BUCKET_IS_EMPTY] =
		"请检查目的 bucket 名字是否为空！"
	BosCliSuggetions[BOSCLI_SRC_BUCKET_DONT_EXIST] =
		"源端 bucket 不存在， 请您检查：\n    1. bucket name 是否拼写正确;\n    2. endpoint 是" +
			"是否正确。"
	BosCliSuggetions[BOSCLI_DST_BUCKET_DONT_EXIST] =
		"目的端 bucket 不存在， 请您检查：\n    1. bucket name 是否拼写正确;\n    2. endpoint 是" +
			"是否正确。"
	BosCliSuggetions[BOSCLI_SRC_OBJECT_IS_EMPTY] =
		"COPY 单个文件时， 源端 object 不能为空！"
	BosCliSuggetions[BOSCLI_BATCH_COPY_DSTOBJECT_END] =
		"源端 object key 必须以 \"/\" 结束!"
	BosCliSuggetions[BOSCLI_BATCH_COPY_SRCOBJECT_END] =
		"COPY单文件的时候，源端 object key 不能以 \"/\"结束！如果您要复制文件夹，请加上 -r。\n" +
			"例如：bcecmd bos cp bos:/bucket1/dir/ bos:/bucket2/dir/ -r"
	BosCliSuggetions[BOSCLI_BATCH_DOWNLOAD_SRCOBJECT_END] =
		"下载单个 Object 的时候，源端 object key 不能以 \"/\"结束！如果您要下载文件夹里的所有" +
			"文件，请加上 -r。\n" +
			"例如：bcecmd bos cp bos:/bucket1/dir/ ./dir/ -r"
	BosCliSuggetions[BOSCLI_COVER_SELF] =
		"您指定的 copy 源地址与目的地址相同， 不需要执行COPY！"
	BosCliSuggetions[BOSCLI_CANT_DOWNLOAD_FILES_TO_FILE] =
		"您指定的本地路径为文件， 请指定为目录地址！"
	BosCliSuggetions[BOSCLI_DIR_IS_NOT_WRITABLE] =
		"您指定的本路径没有可写权限， 请将路径指向的文件夹设置为可写！"
	BosCliSuggetions[BOSCLI_UPLOAD_SRC_CANNT_BE_DIR] =
		"如果您要上传文件夹，请加上 -r。\n例如：bcecmd bso cp ./dir/ bos:/bucket/dir/ -r "
	BosCliSuggetions[BOSCLI_DST_OBJECT_KEY_IS_EMPTY] =
		"请指定上传的文件在BOS上保存的名称!"
	BosCliSuggetions[BOSCLI_UPLOAD_STREAM_TO_DIR] =
		"通过流上传文件时， 你需要指定文件保存的名称!"
	BosCliSuggetions[BOSCLI_RM_DIR_MUST_USE_RECURSIVE] =
		"如果您要删除文件夹请加上  -r" +
			"例如：bcecmd bos rm bos:/bucket -r  或 bcecmd bos rm bos:/bucket/dir/ -r"
	BosCliSuggetions[BOSCLI_EXPIRE_LESS_NONE] =
		"有效时间支持1-43200间的整数。如果需要永久有效的分享链接，可以将有效时间设为-1"
	BosCliSuggetions[BOSCLI_SYNC_EXCLUDE_INCLUDE_TIME_TOG] =
		"exclude-time 和 include-time 不能同时使用！"
	BosCliSuggetions[BOSCLI_SYNC_EXCLUDE_INCLUDE_TOG] =
		"不能同时指定 --exclude 和 --include！"
	BosCliSuggetions[BOSCLI_SYNC_UPLOAD_SRC_MUST_DIR] =
		"Sync源端必须为存在的目录，请检查你输入的路径是否是目录，如果是，请检查您是否拥有读权限！"
	BosCliSuggetions[BOSCLI_SYNC_DOWN_DST_MUST_DIR] =
		"Sync 不支持同步单文件， 你可以使用 cp 上传、下载或者复制单个文件！如果指定的路径为文件" +
			"夹，请您检查您是否有读权限！"
	BosCliSuggetions[BOSCLI_SYNC_LOCAL_TO_LOCAL] =
		"Sync 不支持同步本地文件到本地，你可以使用其他工具（比如 rsync）来本地同步文件！"
	BosCliSuggetions[BOSCLI_SYNC_PROCESS_NUM_LESS_ZERO] =
		"Sync并发数不能小于1， 请你使用 bcecmd -c 重新配置！"
	BosCliSuggetions[BOSCLI_INVALID_SYNY_TYPE] =
		"Sync 类型必需是 'time-size', 'time-size-crc32' 或 'only-crc32'！"
	BosCliSuggetions[BOSCLI_PUT_LIFECYCLE_NO_CONFIG_AND_BUCKET] =
		"请指定要配置生命周期的bucekt name，和生命周期配置文件的地址, 操作示例:\n" +
			"bce bosapi put-lifecycle --lifecycle-config-file lifecycle_bj.json --bucket-name " +
			"bucket1"
	BosCliSuggetions[BOSCLI_PUT_LOG_NO_TARGET_BUCKET] =
		"请指定用于保存日志的bucket (Prefix可选)，操作示例:\n" +
			"    指定Prefix: bce bosapi put-logging --target-bucket bucket2 --target-prefix log " +
			"--bucket-name bucket1 \n" +
			"    不指定Prefix: bce bosapi put-logging --target-bucket bucket2 --bucket-name bucket1"
	BosCliSuggetions[BOSCLI_STORAGE_CLASS_IS_EMPTY] =
		"Storage class（存储类型）为空， 请指定storage class. BOS 当前支持 STANDARD（标准），" +
			"STANDARD_IA（低频）和 COLD（冷存储） 三种存储类型！"
	BosCliSuggetions[BOSCLI_PUT_ACL_CANNED_FILE_SAME_TIME] =
		"不能同时通过canned ACL和ACL文件来设置bucket的ACL"
	BosCliSuggetions[BOSCLI_PUT_ACL_CANNED_DONT_SUPPORT] =
		"Canned ACL 仅支持 private、public-read、public-read-write三种"
	BosCliSuggetions[BOSCLI_PUT_ACL_CANNED_FILE_BOTH_EMPTY] =
		"请指定Bucket 的 ACL配置信息，您可以通过 --canned 指定 canned ACL，或者通过 " +
			"--acl-config-file 从文件中上传ACL"

}

func getCliSuggetions(code BosCliErrorCode, err error) string {
	if msg, ok := BosCliSuggetions[code]; ok {
		return msg
	}
	if msg := boscmd.Suggetions(boscmd.BosErrorCode(code), err); msg != "" {
		return msg
	}
	return ""
}

// Abnormal exist: with code
func bcecliAbnormalExistCode(code BosCliErrorCode) {
	printSuggAndErr(code, nil, "")
}

// Abnormal exist: with code, with error
func bcecliAbnormalExistCodeErr(code BosCliErrorCode, err error) {
	var msg string
	code, msg = getCodeAndMsgFromError(code, err)
	printSuggAndErr(code, err, msg)
}

// Abnormal exist: with code, with error msg
func bcecliAbnormalExistCodeMsg(code BosCliErrorCode, format string, args ...interface{}) {
	printSuggAndErr(code, nil, format, args...)
}

// Abnormal exist: without code, with error
func bcecliAbnormalExistErr(err error) {
	code, msg := getCodeAndMsgFromError(BOSCLI_EMPTY_CODE, err)
	printSuggAndErr(code, err, msg)
}

// Abnormal exist: without code, with error msg
func bcecliAbnormalExistMsg(format string, args ...interface{}) {
	printSuggAndErr(BOSCLI_EMPTY_CODE, nil, format, args...)
}

// Print suggestion and error msg to stdin
func printSuggAndErr(code BosCliErrorCode, err error, format string, args ...interface{}) {
	firstPrint := true

	// print error message
	if format != "" {
		if firstPrint {
			fmt.Printf("\n")
			firstPrint = false
		}
		fmt.Printf("Error: "+format+"\n", args...)
	}

	// get suggetion accroding to error code
	if code != BOSCLI_EMPTY_CODE {
		suggetion := getCliSuggetions(code, err)
		if suggetion != "" {
			if firstPrint {
				fmt.Printf("\n")
				firstPrint = false
			}
			if suggetion != "" {
				fmt.Printf("%s: %s\n", BOSCLI_SUGGETION_PROPMT, suggetion)
			}
		}
	}
	if !firstPrint {
		fmt.Printf("\n")
	}
	os.Exit(1)
}

// Get error code and error msg from error
func getCodeAndMsgFromError(code BosCliErrorCode, err error) (BosCliErrorCode, string) {
	if err == nil {
		return code, ""
	}
	if serverErr, ok := err.(*bce.BceServiceError); ok {
		return BosCliErrorCode(serverErr.Code), serverErr.Message
	} else if clientErr, ok := err.(*bce.BceClientError); ok {
		return BosCliErrorCode(boscmd.LOCAL_BCECLIENTERROR), clientErr.Message
	}
	return code, err.Error()
}

// get error message from error
func getErrorMsg(err error) string {
	if err == nil {
		return ""
	}
	if serverErr, ok := err.(*bce.BceServiceError); ok {
		return serverErr.Message
	}
	return err.Error()
}
