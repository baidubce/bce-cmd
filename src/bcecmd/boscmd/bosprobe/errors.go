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

package bosprobe

import (
	"bcecmd/boscmd"
)

type BosProbeErrorCode string

const (
	PROBE_NOT_CHECK                    = "ProbeNotCheck"
	CODE_SUCCESS                       = "CheckSuccess"
	LOCAL_PROBE_INIT_ERROR             = "ClientProbeInitError"
	LOCAL_ARGS_NO_BUCKET               = "ClientArgsNoBucket"
	LOCAL_ARGS_NO_BUCKET_OR_URL        = "ClientArgsNoBucketOrUrl"
	LOCAL_ARGS_BOTH_URL_OBJECT_EXIST   = "ClientArgsBothUrlObjectExist"
	LOCAL_ARGS_BOTH_URL_BUCKET_EXIST   = "ClientArgsBothUrlBucketExist"
	LOCAL_ARGS_BOTH_URL_ENDPOINT_EXIST = "ClientArgsBothUrlEndpointExist"
	PROBE_INIT_REQUEST_FAILED          = "ProbeInitRequestFailed"
	LOCAL_GENERATETEMPFILE_FAILED      = "ClientGenerateTempFileFailed"
	LOCAL_PROBE_INTERNAL_ERROR         = "ClientProbeInternalError"
	LOCAL_GET_ENDPOINT_BUCKET_FAILED   = "ClientGetEndpointOfBucketFailed"
	LOCAL_PROBE_URL_IS_INVALID         = "ClientProbeUrlIsInvalid"
	SERVER_GET_OBJECT_LIST_DENIED      = "AccessDeniedWhenGetOneObjectName"
	SERVER_RETURN_EMPTY_OBJECT_LIST    = "ServerReturnEmptyObjectList"
)

type NetStatusCode int

const (
	NET_NOT_CHECK          NetStatusCode = 0
	NET_IS_OK              NetStatusCode = 1
	ENDPOINT_CONNECT_ERROR NetStatusCode = 2
	CLIENT_NET_ERROR       NetStatusCode = 3
)

// net error msg
const (
	NET_NOT_CHECK_MSG              = "    NET MSG:\t网络畅通!\n"
	ENDPOINT_CONNECT_ERROR_ERR_MSG = "    NET MSG:\t错误代码 2,  不能连接到endpoint (用户能连上" +
		"公网，但不能连接到 endpoint)!\n"
	CLIENT_NET_ERROR_ERR_MSG = "    NET MSG:\t错误代码 3, 用户网络错误 (用户网络既不能连接" +
		"上公网，也不能连接到endpoint)!\n"
)

const (
	suggestionSubmitOrder = "如果提示没能解决问题，您也可以在百度云的管理控制台创建工单，将BOSPr" +
		"obe生成的日志文件(如果未能够生成日志文件，请将BosProbe的所有输出信息复制到本地文件)反馈" +
		"给我们，我们会尽快处理！"
	suggetionDefault              = "BOSProbe还未添加此错误的处理方法!"
	ENDPOINT_CONNECT_ERROR_PROMPT = "*您的电脑能够连接到公网，但是连接不上 endpoint，请你检查您" +
		"是否输入了正确的endpoint，如果正确，您可以在百度云的管理控制台创建工单，将BOSProbe生成" +
		"的日志文件反馈给我们，我们会尽快处理！\n"
	CLIENT_NET_ERROR_PROMPT = "我们检测到你的网络即不能连接到公网，也不能连接到endpoint，请您检" +
		"查您的网络是否配置正确，另外还请检查您是否输入了正确的endpoint\n"
)

var BosProbesuggetions map[BosProbeErrorCode]string

func init() {
	initSuggetionChinese()
}

func initSuggetionChinese() {
	BosProbesuggetions = make(map[BosProbeErrorCode]string)
	BosProbesuggetions[PROBE_NOT_CHECK] =
		"未执行检测"
	BosProbesuggetions[CODE_SUCCESS] =
		"测试通过"
	BosProbesuggetions[LOCAL_PROBE_INIT_ERROR] =
		"BosProbe 初始化失败！"
	BosProbesuggetions[PROBE_INIT_REQUEST_FAILED] =
		"BosProbe 请求初始化失败， 请检查您使用的命令格式是否正确！"
	BosProbesuggetions[LOCAL_ARGS_NO_BUCKET] =
		"您没有指定bucket 名称，请您在执行命令的时候用 -b bucketName 指定bucket 名称"
	BosProbesuggetions[LOCAL_ARGS_NO_BUCKET_OR_URL] =
		"下载测试时，你需要指定要下载文件所在的bucket name 或 URL， bucket name 通过 -b 指定，" +
			"URL 通过 -f 指定， 不过请注意， 不能同时指定 URL 和 bucket"
	BosProbesuggetions[LOCAL_ARGS_BOTH_URL_BUCKET_EXIST] =
		"下载测试时，你不能同时指定要下载文件所在的bucket name 和 URL，因为这样可能会使" +
			"BOSProbe错误地解析您需要执行的命令."
	BosProbesuggetions[LOCAL_ARGS_BOTH_URL_OBJECT_EXIST] =
		"下载测试时，你不能同时指定要下载文件所在的object name 和 URL，因为这样可能会使" +
			"BOSProbe错误地解析您需要执行的命令."
	BosProbesuggetions[LOCAL_ARGS_BOTH_URL_ENDPOINT_EXIST] =
		"下载测试时，你不能同时指定 endpoint 和 URL，因为这样可能会使BOSProbe错" +
			"误地解析您需要执行的命令."
	BosProbesuggetions[LOCAL_GENERATETEMPFILE_FAILED] =
		"生成临时文件失败， 请检查您是否对当前目录具有写权限后重试！"
	BosProbesuggetions[LOCAL_PROBE_URL_IS_INVALID] =
		"您指定的URL格式不对，请您检查URL是否输入错误！"
	BosProbesuggetions[LOCAL_GET_ENDPOINT_BUCKET_FAILED] =
		"根据Bucket name 获取 endpoint 失败， 请你使用 -e 手动指定 endpoint 后再次执行检测！\n" +
			"BOS访问的域名为：\n" +
			"    区域\t访问的Endpoint\n    北京\tbj.bcebos.com\n" +
			"    广州\tgz.bcebos.com\n    苏州\tsu.bcebos.com\n    香港2区\thk-2.bcebos.com\n"
	BosProbesuggetions[SERVER_RETURN_EMPTY_OBJECT_LIST] =
		"没能够从服务端获取到object name，可能是因为这个bucket 为空。 如果bucket不为空" +
			"，请你用 -o 指定要下载的object name，然后重试。"
	BosProbesuggetions[SERVER_GET_OBJECT_LIST_DENIED] =
		"当从服务端随机获取一个object name的时候，服务端拒绝了请求。可能是因为你没有这个bucket" +
			"的list权限, 或者欠费等其他原因。 为了查看具体原因请你用 -o 指定要下载的object name" +
			"，然后重试。"
}

func probeSuggetions(code BosProbeErrorCode, err error) string {
	if msg, ok := BosProbesuggetions[code]; ok {
		return msg + "\n\n" + suggestionSubmitOrder
	}
	if msg := boscmd.Suggetions(boscmd.BosErrorCode(code), err); msg != "" {
		return msg + "\n\n" + suggestionSubmitOrder
	}
	return suggetionDefault + "\n\n" + suggestionSubmitOrder
}
