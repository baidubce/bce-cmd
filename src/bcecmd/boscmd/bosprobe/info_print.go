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
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

import (
	"utils/net_tools"
	"utils/util"
)

const (
	PING_MAX_COUNT           = "5"
	TRACERT_MAX_HOPS         = "15"
	PROBE_NET_INFO_DELIMITER = ">>>>>>"
	PROBE_CHECK_DELIMITER    = "\n\n************************* %s **************************\n"
)

// replace the real values of ak and sk to '***'
func disturbAkSk(cmd string) string {

	re := regexp.MustCompile(` -a=([^ ]*)`)
	cmd = re.ReplaceAllString(cmd, " -a=***")
	re = regexp.MustCompile(` -a([\S]+)`)
	cmd = re.ReplaceAllString(cmd, " -a***")
	re = regexp.MustCompile(` -a[\s]+([^ ]*)`)
	cmd = re.ReplaceAllString(cmd, " -a ***")
	re = regexp.MustCompile(` --ak=([^ ]*)`)
	cmd = re.ReplaceAllString(cmd, " --ak=***")
	re = regexp.MustCompile(` --ak[\s]+([^ ]*)`)
	cmd = re.ReplaceAllString(cmd, " --ak ***")
	re = regexp.MustCompile(` -s=([^ ]*)`)
	cmd = re.ReplaceAllString(cmd, " -s=***")
	re = regexp.MustCompile(` -s([\S]+)`)
	cmd = re.ReplaceAllString(cmd, " -s***")
	re = regexp.MustCompile(` -s[\s]+([^ ]*)`)
	cmd = re.ReplaceAllString(cmd, " -s ***")
	re = regexp.MustCompile(` --sk=([^ ]*)`)
	cmd = re.ReplaceAllString(cmd, " --sk=***")
	re = regexp.MustCompile(` --sk[\s]+([^ ]*)`)
	cmd = re.ReplaceAllString(cmd, " --sk ***")

	return cmd
}

// get the net status of client
func NetStatusReport(endpoint, authAddress string) NetStatusCode {
	pLog.logging(PROBE_CHECK_DELIMITER, "Net Status")
	// ping endpoint
	pLog.logging("%s ping %s\n", PROBE_NET_INFO_DELIMITER, endpoint)
	pingEndpointOut, pingEndpiontErr := net_tools.Ping(&endpoint, PING_MAX_COUNT)
	if len(pingEndpointOut) != 0 {
		pLog.logging("%s", pingEndpointOut)
	}
	if pingEndpiontErr != nil {
		pLog.logging("%s", pingEndpiontErr)
	}
	pLog.logging("\n\n")

	//ping auth address
	pLog.logging("%s ping %s\n", PROBE_NET_INFO_DELIMITER, authAddress)
	pingAuthOut, pingAuthErr := net_tools.Ping(&authAddress, PING_MAX_COUNT)
	if len(pingAuthOut) != 0 {
		pLog.logging("%s", pingAuthOut)
	}
	if pingAuthErr != nil {
		pLog.logging("%s", pingAuthErr)
	}
	pLog.logging("\n\n")

	//nslookup endpoint
	pLog.logging("%s nslookup %s\n", PROBE_NET_INFO_DELIMITER, endpoint)
	nsEndpointOut, nsEndpointErr := net_tools.Nslookup(&endpoint)
	if len(nsEndpointOut) != 0 {
		pLog.logging("%s", nsEndpointOut)
	}
	if nsEndpointErr != nil {
		pLog.logging("%s", pingAuthErr)
	}
	pLog.logging("\n\n")

	//nslookup auth address
	pLog.logging("%s nslookup %s\n", PROBE_NET_INFO_DELIMITER, authAddress)
	nsAuthOut, nsAuthErr := net_tools.Nslookup(&authAddress)
	if len(nsAuthOut) != 0 {
		pLog.logging("%s", nsAuthOut)
	}
	if nsAuthErr != nil {
		pLog.logging("%s", nsAuthErr)
	}
	pLog.logging("\n\n")

	//tracert endpoint
	pLog.logging("%s tracert %s\n", PROBE_NET_INFO_DELIMITER, endpoint)
	traceEpOut, traceEpErr := net_tools.Tracert(&endpoint, TRACERT_MAX_HOPS)
	if len(traceEpOut) != 0 {
		pLog.logging("%s", traceEpOut)
	}
	if traceEpErr != nil {
		pLog.logging("%s", traceEpErr)
	}
	pLog.logging("\n\n")

	if pingEndpiontErr == nil {
		return NET_IS_OK
	} else if pingAuthErr == nil {
		return ENDPOINT_CONNECT_ERROR
	} else {
		return CLIENT_NET_ERROR
	}
}

// save basic info of client to log file
// contain: system info, exec time, command
func getBasicInfo() error {
	cmd := strings.Join(os.Args[:], " ")
	cmd = disturbAkSk(cmd)

	pLog.logging(PROBE_CHECK_DELIMITER, "Basic Info")
	pLog.logging("    TIME   :\t%s\n", time.Now())
	pLog.logging("    SYSTEM :\t%s\n", runtime.GOOS)
	pLog.logging("    COMMAND:\t%s\n\n", cmd)
	return nil
}

func printRequestCheckInfo(code BosProbeErrorCode) {
	msg := fmt.Sprintf(PROBE_CHECK_DELIMITER, "参数检查结果")
	msg += "    检测"
	if code != CODE_SUCCESS {
		msg += "[失败]\n"
		msg += fmt.Sprintf("    失败代码：%s\n", code)
	} else {
		msg += "[成功]\n"
	}
	pLog.logging(msg)
}

// if success: print the info of file
func printCheckDetail(propmt string, objectInfo *objectDetail) {
	detail := fmt.Sprintf(PROBE_CHECK_DELIMITER, propmt)
	detail += fmt.Sprintf("    OBJECT NAME:\t%s\n", objectInfo.objectName)
	detail += fmt.Sprintf("    FILE SIZE  :\t%d\n", objectInfo.objectSize)
	detail += fmt.Sprintf("    USED TIME  :\t%d ms\n", objectInfo.useTimeMillisecond)
	detail += fmt.Sprintf("    SPEED      :\t%.2f MB/s\n",
		float64(objectInfo.objectSize)/1048.576/float64(objectInfo.useTimeMillisecond))
	pLog.Tlogging(detail)
}

// print result of bosprobe: success or fail
func printCheckResult(propmt string, netStatus NetStatusCode, checkCode BosProbeErrorCode) {
	retMsg := fmt.Sprintf(PROBE_CHECK_DELIMITER, "测试结果")
	if netStatus == NET_IS_OK {
		retMsg += "    网络测试结果 [网络畅通]\n"
	} else if netStatus == NET_NOT_CHECK {
		retMsg += "    网络未测试\n"
	} else {
		retMsg += "    网络测试结果 [网络不通]\n"
	}

	retMsg += fmt.Sprintf("    %s测试结果", propmt)
	if checkCode == CODE_SUCCESS {
		retMsg += fmt.Sprintf(" [%s成功]\n", propmt)
	} else {
		retMsg += fmt.Sprintf(" [%s失败]\n", propmt)
	}
	pLog.logging(retMsg)
}

func PrintNetStatus(netCode NetStatusCode) {
	if netCode == NET_NOT_CHECK {
		return
	}
	netMsg := fmt.Sprintf(PROBE_CHECK_DELIMITER, "网络信息")
	if netCode == NET_IS_OK {
		netMsg += NET_NOT_CHECK_MSG
	} else if netCode == ENDPOINT_CONNECT_ERROR {
		netMsg += ENDPOINT_CONNECT_ERROR_ERR_MSG
	} else if netCode == CLIENT_NET_ERROR {
		netMsg += CLIENT_NET_ERROR_ERR_MSG
	}
	pLog.logging(netMsg)
}

func PrintFailedDetail(checkErrorCode BosProbeErrorCode, err error, printToCoTerminal bool) {

	if checkErrorCode == CODE_SUCCESS || checkErrorCode == PROBE_NOT_CHECK {
		return
	}

	errMsg := fmt.Sprintf(PROBE_CHECK_DELIMITER, "错误信息")
	if checkErrorCode != "" {
		errMsg += fmt.Sprintf("    ERROR CODE:\t%s\n", checkErrorCode)
	}
	if err != nil {
		errMsg += fmt.Sprintf("    ERROR MSG :\t%s\n", err)
	}

	if printToCoTerminal {
		pLog.Tlogging(errMsg)
	} else {
		pLog.logging(errMsg)
	}
}

func PrintDebugInfo(resp *serviceResp) {
	if resp == nil {
		return
	}
	pLog.logging(PROBE_CHECK_DELIMITER, "调试信息")
	if resp.requestId != "" {
		pLog.logging("    Request ID:\t%s\n", resp.requestId)
	}
	if resp.debugId != "" {
		pLog.logging("    Debug   ID:\t%s\n", resp.debugId)
	}
}

// if failed: giving suggestions
func giveSuggestion(netCode NetStatusCode, checkErrorCode BosProbeErrorCode, err error) {
	suggMsg := fmt.Sprintf(PROBE_CHECK_DELIMITER, "建    议")
	if netCode == ENDPOINT_CONNECT_ERROR {
		suggMsg += ENDPOINT_CONNECT_ERROR_PROMPT
	} else if netCode == CLIENT_NET_ERROR {
		suggMsg += CLIENT_NET_ERROR_PROMPT
	} else {
		suggMsg += probeSuggetions(checkErrorCode, err) + "\n"
	}
	pLog.Tlogging(suggMsg)
}

func PrintLogInfo() {
	//get path of log file
	if pLog.createLogFileError {
		return
	}
	logPath, err := os.Getwd()
	if err == nil {
		if !strings.HasSuffix(logPath, util.OsPathSeparator) {
			logPath += util.OsPathSeparator
		}
	} else {
		logPath = "./"
	}
	fmt.Printf("\n此次测试的日志保存于: %s%s\n\n", logPath, pLog.logName)
}
