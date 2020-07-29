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

// Constants of BOSCLI.

package boscli

const (
	DEFAULT_STORAGE_CLASS = "STANDARD"

	BCE_CLI_AGENT  = "bce-go-cli"
	HTTP_PROTOCOL  = "http://"
	HTTPS_PROTOCOL = "https://"

	GOOS_WINDOWS_DEF       = "windows"
	BOS_PATH_PREFIX        = "bos:/"
	BOS_PATH_PREFIX_DOUBLE = "bos://"
	BOS_TIME_FORMT         = "2006-01-02T15:04:05Z"
	LOCAL_TIME_FROMT       = "2006-01-02 15:04:05"
	BOS_HTTP_TIME_FORMT    = "Mon, 02 Jan 2006 15:04:05 MST"

	SIGNED_URL_EXPIRE_TIME      = 1800
	GAP_GET_OBJECT_INFO_AGAIN   = 60 //60s
	MAX_PARTS                   = 10000
	MAX_STREAM_UPLOAD_SIZE      = 5 << 30 // 5G
	STREAM_DOWNLOAD_BUF_SIZE    = 2 << 20
	SYNC_COMPARATOR_TIME_OUT    = 36000 * 1000 // 10 hours
	GET_NET_LOCAL_FILE_TIME_OUT = 36000 * 1000 // 10 hours
	MULTI_UPLOAD_MAX_FILE_SIZE  = 5 << 40      // 5T
	MULTI_UPLOAD_THRESHOLD      = 32 << 20     // 32M
	PART_SIZE_BASE              = 10 << 20     // 10M
	MULTI_COPY_THRESHOLD        = 100 << 20    // 100M
	MULTI_COPY_PART_SIZE        = 50 << 20     // 50M

	MULTI_DOWNLOAD_THRESHOLD = 100 << 20 // 32M
)

// sync op constants
const (
	// op type
	OPERATE_CMD_COPY    = "operateCmdCopy"
	OPERATE_CMD_DELETE  = "operateCmdDelete"
	OPERATE_CMD_NOTHING = "operateCmdNothing"

	SYNC_OP_COPY     = "Copy"
	SYNC_OP_DOWNLOAD = "Download"
	SYNC_OP_UPLOAD   = "Upload"
	SYNC_OP_DELETE   = "Delete" // delete local file
	SYNC_OP_REMOVE   = "Remove" // delete bos object
	SYNC_OP_ERROR    = "Error"

	IS_BOS         = "bos"
	IS_LOCAL       = "local"
	BOS_TO_BOS     = "bosbos"
	BOS_TO_LOCAL   = "boslocal"
	LOCAL_TO_LOCAL = "locallocal"
	LOCAL_TO_BOS   = "localbos"
)
