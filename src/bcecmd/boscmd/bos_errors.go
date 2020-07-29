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

package boscmd

import (
	"strings"
)

type BosErrorCode string

const (
	LOCAL_PATH_NOT_EXIST             = "ClientPathNotExist"
	LOCAL_FILE_NOT_EXIST             = "ClientLocalFileNotExist"
	LOCAL_OPEN_FILE_FAILED           = "ClientOpenLocalFileFailed"
	LOCAL_GET_REAL_OF_SYMLINK_FAILED = "ClientFollowSymlinkFailed"
	LOCAL_INIT_BOSCLIENT_FAILED      = "ClientInitBosClientFailed"
	LOCAL_BCECLIENTERROR             = "CleintBceError"
	CODE_SIGNATURE_DOES_NOT_MATCH    = "SignatureDoesNotMatch"
	CODE_ACCESS_DENIED               = "AccessDenied"
	CODE_ACCOUNT_OVERDUE             = "AccountOverdue"
	CODE_ACCESS_DENIED_BY_SOURCE_URL = "AccessDeniedBySourceUrl"
	CODE_FETCH_OBJECT_FAILED         = "FetchObjectFailed"
	CODE_INVALID_ACCESS_KEY_ID       = "InvalidAccessKeyId"
	CODE_INVALID_BUCKET_NAME         = "InvalidBucketName"
	CODE_INVALID_OBJECT_NAME         = "InvalidObjectName"
	CODE_NO_SUCH_BUCKET              = "NoSuchBucket"
	CODE_NO_SUCH_KEY                 = "NoSuchKey"
	CODE_SERVICE_UNAVAILABLE         = "ServiceUnavailable"
	CODE_SLOW_DOWN                   = "SlowDown"
	CODE_INTERNAL_ERROR              = "InternalError"
	CODE_CANT_GET_ONE_OBJECT_NAME    = "ServerCantGetOneObject"
	CODE_REQUEST_IS_EXPIRED          = "RequestExpired"
	CODE_ENTITY_TOO_LARGE            = "EntityTooLarge"
	CODE_ENTITY_TOO_SMALL            = "EntityTooSmall"
	CODE_INVALID_HTTP_AUTH_HEADER    = "InvalidHTTPAuthHeader"
	CODE_INVALID_URL                 = "InvalidURI"
	CODE_REQUEST_TIMEOUT             = "RequestTimeout"
	CODE_BUCKET_ALREADY_EXISTS       = "BucketAlreadyExists"
	CODE_BUCKET_NOT_EMPTY            = "BucketNotEmpty"
	CODE_INVALID_ARGUMENT            = "InvalidArgument"
	CODE_NO_SUCH_UPLOAD              = "NoSuchUpload"
	CODE_INVALID_PART                = "InvalidPart"
	CODE_INVALID_PART_ORDER          = "InvalidPartOrder"
)

const (
	suggetionAccessDeniedOne = "您的请求频率过高， BOS限制了您的部分请求， 请您降低您的请求频率！"
	suggetionAccessDeniedTwo = "请求被服务端拒绝啦！请确认如下几点：\n    1. 您对要操作的bucket是" +
		"否有权限， 如果Bucket不是公共读或公共写，请您检查您的请求是否设置了AK/SK；\n" +
		"    2. 您的账户是否欠费；\n    3. 参数是否正确；\n    4. 如果你是通过URL下载，" +
		"请检查URL是否正确，是否带了认证信息(如果是公共读可以不带认证信息)。\n"
)

var BosSuggetions map[BosErrorCode]string

func init() {
	initSuggetionChinese()
}

func initSuggetionChinese() {
	BosSuggetions = make(map[BosErrorCode]string)
	BosSuggetions[LOCAL_PATH_NOT_EXIST] =
		"您指定的本地路径不存在，请检查您指定的路径是否存在，如果存在，请检查您是否有权限读！"
	BosSuggetions[LOCAL_FILE_NOT_EXIST] =
		"您指定的本地文件不存在，请检查您指定的路径是否存在, 如果存在,请检查是否有权限读取！"
	BosSuggetions[LOCAL_OPEN_FILE_FAILED] =
		"打开文件失败， 请检查你指定的本地文件是否存在， 然后确定对此文件是否具有读权限！"
	BosSuggetions[LOCAL_GET_REAL_OF_SYMLINK_FAILED] =
		"读取软链接信息失败。 1. 请检查您是否有权限读取此文件。 2. 请检查软链接文件指向的真实文" +
			"件是否存在！"
	BosSuggetions[LOCAL_INIT_BOSCLIENT_FAILED] =
		"BOS Client 初始化失败, 请您检查你输入的参数是否正确！"
	BosSuggetions[LOCAL_BCECLIENTERROR] =
		"客户端执行失败，请您检查你输入的参数是否正确，然后重试！"
	BosSuggetions[CODE_SIGNATURE_DOES_NOT_MATCH] =
		"签名不匹配，请您检查您提供的AK/SK是否正确！"
	BosSuggetions[CODE_ACCOUNT_OVERDUE] =
		"您的请求被服务端拒绝了，因为您的账户已欠费！"
	BosSuggetions[CODE_ACCESS_DENIED_BY_SOURCE_URL] =
		"您配置了镜像回源，但是当我们获取您指定的url的时候，源站拒绝了我们的请求， 请确认您对" +
			"源站是否有权限.\n说明:\n    1. 目前不支持对图片服务相关GetObject请求进行镜像回源。\n" +
			"    2. BOS在进行镜像回源时，不会携带原请求中的QueryString。\n"
	BosSuggetions[CODE_FETCH_OBJECT_FAILED] =
		"从源站获取文件失败，请您检测您设置的参数!"
	BosSuggetions[CODE_INVALID_ACCESS_KEY_ID] =
		"您提供的BOS access key ID 不存在!"
	BosSuggetions[CODE_INVALID_BUCKET_NAME] =
		"你指定的bucket name格式不正确， 请检查bucket name是否写对！"
	BosSuggetions[CODE_INVALID_OBJECT_NAME] =
		"您指定的object name 太长了, 请减少object name 的长度 （最长为1024个字节）"
	BosSuggetions[CODE_NO_SUCH_BUCKET] =
		"请检查您指定的 bucket 是否存在，如果存在，请您确认bucket name是否拼写正确、endpoint" +
			"是否正确!"
	BosSuggetions[CODE_NO_SUCH_KEY] =
		"您指定的object在BOS中不存在， 请您检查object name是否拼写正确! 如果指定的 object key " +
			"为目录， 请给object key 加上后缀 \"/\"。"
	BosSuggetions[CODE_SERVICE_UNAVAILABLE] =
		"您的请求频率过高，请您降低您的请求频率!"
	BosSuggetions[CODE_SLOW_DOWN] =
		"您的请求频率过高，请您降低您的请求频率!"
	BosSuggetions[CODE_INTERNAL_ERROR] =
		"BOS端发生错误， 请您重试您的操作!"
	BosSuggetions[CODE_REQUEST_IS_EXPIRED] =
		"您请求的已过期， 可能是因为你使用的URL配置了过期时间， 而您发送请求时， URL已经" +
			"过期了，你可以重新生成URL。"
	BosSuggetions[CODE_ENTITY_TOO_LARGE] =
		"错误的原因是您上传的数据大于限制。单个文件上传最大限制为5TB， 如果你的文件大于5GB，" +
			"请使用三步上传."
	BosSuggetions[CODE_ENTITY_TOO_SMALL] =
		"上传的数据小于限制."
	BosSuggetions[CODE_INVALID_HTTP_AUTH_HEADER] =
		"Authorization头域格式错误, 请你检查URL是否完整。"
	BosSuggetions[CODE_INVALID_URL] =
		"URI形式不正确， 请您检查URL输入是否正确."
	BosSuggetions[CODE_REQUEST_TIMEOUT] =
		"请求超时"
	BosSuggetions[CODE_BUCKET_ALREADY_EXISTS] =
		"您要创建的 Bucket Name 已经被使用了， 请选择一个不同的 Bucket Name， 然后重试！"
	BosSuggetions[CODE_BUCKET_NOT_EMPTY] =
		"你要删除的Bucket 不为空!"
	BosSuggetions[CODE_INVALID_ARGUMENT] =
		"请检查你的输入的参数是否正确！"
}

func Suggetions(code BosErrorCode, err error) string {
	if msg, ok := BosSuggetions[code]; ok {
		return msg
	}
	if code == CODE_ACCESS_DENIED && err != nil {
		if strings.Contains(err.Error(), "request rate is too high") {
			return suggetionAccessDeniedOne
		} else {
			return suggetionAccessDeniedTwo
		}
	}
	return ""
}
