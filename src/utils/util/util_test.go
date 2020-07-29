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

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	userHomeDir string
)

func init() {
	InitTestFiles()
	var err error
	if userHomeDir, err = GetHomeDirOfUser(); err != nil {
		fmt.Println("Error: init get home dir", err.Error())
		os.Exit(1)
	}
}

func TestGetRandomString(t *testing.T) {
	testCases := []int64{1, 2, 100, 1000, 10000, 10000, 1000000}
	temp := ""
	for i, fileSize := range testCases {
		ret := GetRandomString(fileSize)
		ExpectEqual("util TestGetRandomString", i, t.Errorf, fileSize, len(ret))
		if ret == temp {
			t.Errorf("two random string is the same: %s \n %s\n", temp, ret)
		}
		temp = ret
	}
}

type TestGetHostFromUrlType struct {
	url   string
	host  string
	isErr bool
}

func TestGetHostFromUrl(t *testing.T) {
	testCases := []TestGetHostFromUrlType{
		TestGetHostFromUrlType{
			url:   "www.baidu.com/ttest?quer=1",
			isErr: true,
		},
		TestGetHostFromUrlType{
			url:   "baidu.com/ttest?quer=1",
			isErr: true,
		},
		TestGetHostFromUrlType{
			url:  "http://baidu.com/ttest?quer=1",
			host: "baidu.com",
		},
		TestGetHostFromUrlType{
			url:  "http://www.baidu.com/ttest?quer=1",
			host: "www.baidu.com",
		},
		TestGetHostFromUrlType{
			url:  "https://www.baidu.com/ttest?quer=1",
			host: "www.baidu.com",
		},
		TestGetHostFromUrlType{
			url:   "htt/ww.baidu.com/ttest?quer=1",
			isErr: true,
		},
		TestGetHostFromUrlType{
			url:   "httww",
			isErr: true,
		},
	}
	for i, tCase := range testCases {
		ret, err := GetHostFromUrl(tCase.url)
		ExpectEqual("util GetHostFromUrl I", i, t.Errorf, tCase.isErr, err != nil)
		if err == nil {
			t.Logf("host of %s is %s", tCase.url, ret)
		}
		if tCase.isErr == false {
			ExpectEqual("util GetHostFromUrl II", i, t.Errorf, tCase.host, ret)
		}
	}

}

type pathExistType struct {
	path  string
	exist bool
}

func TestDoesPathExist(t *testing.T) {
	testCases := []pathExistType{
		//0
		pathExistType{
			path:  "~/xxxx",
			exist: false,
		},
		pathExistType{
			path:  "~/debug_bos",
			exist: false,
		},
		//2
		pathExistType{
			path:  "./util.go",
			exist: true,
		},
		pathExistType{
			path:  "../main",
			exist: false,
		},
		//4
		pathExistType{
			path:  "../xxmain",
			exist: false,
		},
		pathExistType{
			path:  "/xxmain",
			exist: false,
		},
		//6
		pathExistType{
			path:  "/root/install.log",
			exist: false,
		},
		pathExistType{
			path:  "/root/ld_trust_for_init",
			exist: false,
		},
	}
	for i, tCase := range testCases {
		ret := DoesPathExist(tCase.path)
		ExpectEqual("util DoesPathExist", i, t.Errorf, tCase.exist, ret)
	}
}

func TestDoesFileExist(t *testing.T) {
	testCases := []pathExistType{
		//0
		pathExistType{
			path:  "~/xxxx",
			exist: false,
		},
		pathExistType{
			path:  "~/debug_bos/BOSCLI.md",
			exist: false,
		},
		//2
		pathExistType{
			path:  "~/debug_bos",
			exist: false,
		},
		pathExistType{
			path:  "./util.go",
			exist: true,
		},
		//4
		pathExistType{
			path:  "../main",
			exist: false,
		},
		pathExistType{
			path:  "../xxmain",
			exist: false,
		},
		//6
		pathExistType{
			path:  "/xxmain",
			exist: false,
		},
		pathExistType{
			path:  "/root/install.log",
			exist: false,
		},
	}
	for i, tCase := range testCases {
		ret := DoesFileExist(tCase.path)
		ExpectEqual("util DoesFileExist", i, t.Errorf, tCase.exist, ret)
	}
}

func TestDoesDirExist(t *testing.T) {
	testCases := []pathExistType{
		//0
		pathExistType{
			path:  "~/xxxx",
			exist: false,
		},
		pathExistType{
			path:  "~/debug_bos",
			exist: false,
		},
		//2
		pathExistType{
			path:  "./util.go",
			exist: false,
		},
		pathExistType{
			path:  "../main",
			exist: false,
		},
		//4
		pathExistType{
			path:  "../xxmain",
			exist: false,
		},
		pathExistType{
			path:  "/xxmain",
			exist: false,
		},
		//6
		pathExistType{
			path:  "/root/install.log",
			exist: false,
		},
		pathExistType{
			path:  "/root",
			exist: true,
		},
		//8
		pathExistType{
			path:  "/root/ld_trust_for_init",
			exist: false,
		},
	}
	for i, tCase := range testCases {
		ret := DoesDirExist(tCase.path)
		ExpectEqual("util DoesDirExist", i, t.Errorf, tCase.exist, ret)
	}
}

func TestIsDirWritable(t *testing.T) {
	testCases := []pathExistType{
		pathExistType{
			path:  "~/xxxx",
			exist: false,
		},
		pathExistType{
			path:  "./",
			exist: true,
		},
		pathExistType{
			path:  "./util.go",
			exist: false,
		},
		pathExistType{
			path:  "../main",
			exist: false,
		},
		pathExistType{
			path:  "../xxmain",
			exist: false,
		},
		pathExistType{
			path:  "/xxmain",
			exist: false,
		},
		pathExistType{
			path:  "/root",
			exist: false,
		},
	}
	for i, tCase := range testCases {
		ret := IsDirWritable(tCase.path)
		ExpectEqual("util IsDirWritable", i, t.Errorf, tCase.exist, ret)
	}
}

func TestIsFileWritable(t *testing.T) {
	testCases := []pathExistType{
		//0
		pathExistType{
			path:  "~/xxxx",
			exist: false,
		},
		pathExistType{
			path:  "~/debug_bos/BOSCLI.md",
			exist: false,
		},
		//2
		pathExistType{
			path:  "~/debug_bos",
			exist: false,
		},
		pathExistType{
			path:  "./test_tools.go",
			exist: true,
		},
		//4
		pathExistType{
			path:  "../main",
			exist: false,
		},
		pathExistType{
			path:  "../xxmain",
			exist: false,
		},
		//6
		pathExistType{
			path:  "/xxmain",
			exist: false,
		},
		pathExistType{
			path:  "/root/install.log",
			exist: false,
		},
	}
	for i, tCase := range testCases {
		var omtime int64
		if DoesFileExist(tCase.path) {
			fstat, err := os.Stat(tCase.path)
			if err != nil {
				t.Errorf("can't get info of %s", tCase.path)
			}
			omtime = fstat.ModTime().Unix()
		}
		ret := IsFileWritable(tCase.path)
		if ret {
			fstat, err := os.Stat(tCase.path)
			if err != nil {
				t.Errorf("can't get info of %s", tCase.path)
			}
			mtime := fstat.ModTime().Unix()
			if mtime != omtime {
				t.Errorf("old mtime of file %s is %d new is %d", tCase.path, omtime, mtime)
			} else {
				t.Logf("%s old mtime %d new mtime %d", tCase.path, omtime, mtime)
			}
		}
		ExpectEqual("util IsFileWritable", i, t.Errorf, tCase.exist, ret)
	}
}

func TestIsSymbolicLink(t *testing.T) {
	newname := "./test_tools_slink.go"
	oldname := "./test_tools.go"
	err := os.Symlink(oldname, newname)
	if err != nil {
		t.Errorf("create synlik  %s to %s failed", newname, oldname)
		return
	}
	isLink, err := IsSymbolicLink(newname)
	if err != nil {
		t.Errorf("check synlik  %s failed", newname)
	} else {
		ExpectEqual("util TestIsSymbolicLink", 1, t.Errorf, true, isLink)
		os.Remove(newname)
	}

	isLink, err = IsSymbolicLink(oldname)
	if err != nil {
		t.Errorf("check synlik  %s failed", oldname)
	} else {
		ExpectEqual("util TestIsSymbolicLink", 1, t.Errorf, false, isLink)
	}

	newname = "./test_tools_link.go"
	err = os.Link(oldname, newname)
	if err != nil {
		t.Errorf("create hard link  %s to %s failed", newname, oldname)
		return
	}
	isLink, err = IsSymbolicLink(newname)
	if err != nil {
		t.Errorf("check synlik  %s failed", newname)
	} else {
		ExpectEqual("util TestIsSymbolicLink", 1, t.Errorf, false, isLink)
		os.Remove(newname)
	}

	newname = "./xxx3432"
	_, err = IsSymbolicLink(newname)
	ExpectEqual("util TestIsSymbolicLink", 1, t.Errorf, false, err == nil)
}

func TestTryMkdir(t *testing.T) {
	os.Create("./cover.html")
	testCases := []pathExistType{
		//0
		pathExistType{
			path:  "./xxxx",
			exist: true,
		},
		pathExistType{
			path:  "./cover.html",
			exist: false,
		},
		//2
		pathExistType{
			path:  userHomeDir + "/xxmain",
			exist: true,
		},
		pathExistType{
			path:  "/xxmain",
			exist: false,
		},
		//4
		pathExistType{
			path:  "/root/test",
			exist: false,
		},
	}

	for i, tCase := range testCases {
		err := TryMkdir(tCase.path)
		ExpectEqual("util TryMkdir", i, t.Errorf, tCase.exist, err == nil)
	}
}

func TestGetSizeOfFile(t *testing.T) {
	testCases := []pathExistType{
		pathExistType{
			path:  "./xxxx",
			exist: false,
		},
		pathExistType{
			path:  "./util.go",
			exist: true,
		},
		pathExistType{
			path:  userHomeDir + "/debug_bos",
			exist: false,
		},
		pathExistType{
			path:  "/xxmain",
			exist: false,
		},
		pathExistType{
			path:  "/root/install.log",
			exist: false,
		},
		pathExistType{
			path:  "~/debug_bos/autoconf-2.69.tar.gz",
			exist: false,
		},
	}

	for i, tCase := range testCases {
		ret, err := GetSizeOfFile(tCase.path)
		ExpectEqual("util GetSizeOfFile I", i, t.Errorf, tCase.exist, err == nil)
		if err == nil {
			t.Logf("file %s size %d", tCase.path, ret)
		}
		if tCase.exist == true {
			ExpectEqual("util GetSizeOfFile II", i, t.Errorf, true, ret > 0)
		}
	}
}

// the files are:
// ./a/
// ./a/b
// ./a-b
// ./a_b
// ./aA
// ./aa
// ab

const (
	SORT_DIR_TEST_PATH = "./sortTestDirPath"
)

func generateFilesForReadSortedDirName() error {
	if err := TryMkdir(SORT_DIR_TEST_PATH); err != nil {
		return err
	}
	if err := TryMkdir(SORT_DIR_TEST_PATH + "/a"); err != nil {
		return err
	}
	if err := CreateFileWithSize(SORT_DIR_TEST_PATH+"/a-b", 1); err != nil {
		return err
	}
	if err := CreateFileWithSize(SORT_DIR_TEST_PATH+"/a_b", 1); err != nil {
		return err
	}
	if err := TryMkdir(SORT_DIR_TEST_PATH + "/a/b"); err != nil {
		return err
	}
	if err := CreateFileWithSize(SORT_DIR_TEST_PATH+"/aA", 1); err != nil {
		return err
	}
	if err := CreateFileWithSize(SORT_DIR_TEST_PATH+"/ab", 1); err != nil {
		return err
	}
	if err := TryMkdir(SORT_DIR_TEST_PATH + "/a/c"); err != nil {
		return err
	}
	return nil
}

type dirNamesType struct {
	path      string
	listNum   int
	isErr     bool
	filesList []string
}

func TestReadSortedDirNames(t *testing.T) {
	err := generateFilesForReadSortedDirName()
	if err != nil {
		t.Errorf("ReadSortedDirNames failed generate test files %s ", SORT_DIR_TEST_PATH)
		return
	}
	os.RemoveAll("./test_file")
	TryMkdir("./test_file")

	testCases := []dirNamesType{
		// 0
		dirNamesType{
			path:    "./xxxx",
			listNum: 0,
			isErr:   false,
		},
		dirNamesType{
			path:    "./util.go",
			listNum: 1,
			isErr:   true,
		},
		// 2
		dirNamesType{
			path:  "/root/install.log",
			isErr: true,
		},
		dirNamesType{
			path:  "./test_file",
			isErr: false,
		},
		// 4
		dirNamesType{
			path:      SORT_DIR_TEST_PATH,
			isErr:     false,
			listNum:   5,
			filesList: []string{"a-b", "a/", "aA", "a_b", "ab"},
		},
		// 5
		dirNamesType{
			path:      SORT_DIR_TEST_PATH + "/a",
			isErr:     false,
			listNum:   2,
			filesList: []string{"b", "c"},
		},
		// 6
		dirNamesType{
			path:      SORT_DIR_TEST_PATH + "/a/c",
			isErr:     false,
			listNum:   0,
			filesList: []string{},
		},
	}

	for i, tCase := range testCases {
		ret, err := ReadSortedDirNames(tCase.path)
		ExpectEqual("util ReadSortedDirNames ", i, t.Errorf, tCase.isErr, err != nil)
		if err != nil {
			t.Logf("%s error: %s", tCase.path, err)
		}
		if tCase.isErr == false {
			ExpectEqual("util ReadSortedDirNames I", i, t.Errorf, tCase.listNum, len(ret))
			t.Logf("id %d case %v\n", i, tCase.filesList)
			t.Logf("id %d ret %v\n", i, ret)
			if len(tCase.filesList) != 0 {
				for i, val := range ret {
					val = strings.Replace(val, OsPathSeparator, BOS_PATH_SEPARATOR, -1)
					if val != tCase.filesList[i] {
						t.Errorf("id %d util ReadSortedDirNames II '%s' != '%s' ", i, val, tCase.filesList[i])
						break
					}
				}
			}
		}
	}
	os.RemoveAll(SORT_DIR_TEST_PATH)
}

func TestGetHomeDirOfUser(t *testing.T) {
	_, err := GetHomeDirOfUser()
	ExpectEqual("util GetHomeDirOfUser ", 1, t.Errorf, true, err == nil)
}

type TranUTCTimeStringToTimeStampType struct {
	utcString string
	timeForm  string
	timeStamp int64
	isErr     bool
}

func TestTranUTCTimeStringToTimeStamp(t *testing.T) {
	testCases := []TranUTCTimeStringToTimeStampType{
		TranUTCTimeStringToTimeStampType{
			utcString: "2017-12-05T12:48:30Z",
			timeForm:  "2006-01-02T15:04:05Z",
			timeStamp: 1512478110,
		},
		TranUTCTimeStringToTimeStampType{
			utcString: "2017-12-05T12:48:30Z",
			timeForm:  "Mon, 02 Jan 2006 15:04:05 MST",
			isErr:     true,
		},
		TranUTCTimeStringToTimeStampType{
			utcString: "Wed, 29 Nov 2017 17:18:59 GMT",
			timeForm:  "Mon, 02 Jan 2006 15:04:05 MST",
			timeStamp: 1511975939,
		},
	}
	for i, tCase := range testCases {
		ret, err := TranUTCTimeStringToTimeStamp(tCase.utcString, tCase.timeForm)
		ExpectEqual("util TranUTCTimeStringToTimeStamp ", i, t.Errorf, tCase.isErr, err != nil)
		if err != nil {
			t.Logf("%s error: %s", tCase.utcString, err)
		}
		if tCase.isErr == false {
			ExpectEqual("util TranUTCTimeStringToTimeStamp I", i, t.Errorf, tCase.timeStamp,
				ret)
		}
	}
}

type TestTranUTCtoLocalTimeType struct {
	utcString   string
	timeForm    string
	newTimeForm string
	localString string
	isErr       bool
}

func TestTranUTCtoLocalTime(t *testing.T) {
	testCases := []TestTranUTCtoLocalTimeType{
		TestTranUTCtoLocalTimeType{
			utcString:   "2017-12-05T12:48:30Z",
			timeForm:    "2006-01-02T15:04:05Z",
			newTimeForm: "2006-01-02 15:04:05",
			localString: "2017-12-05 20:48:30",
		},
		TestTranUTCtoLocalTimeType{
			utcString:   "2017-12-05T12:48:30Z",
			timeForm:    "Mon, 02 Jan 2006 15:04:05 MST",
			newTimeForm: "2006-01-02 15:04:05",
			isErr:       true,
		},
	}
	for i, tCase := range testCases {
		ret, err := TranUTCtoLocalTime(tCase.utcString, tCase.timeForm, tCase.newTimeForm)
		ExpectEqual("util TranUTCtoLocalTime ", i, t.Errorf, tCase.isErr, err != nil)
		if err != nil {
			t.Logf("%s error: %s", tCase.utcString, err)
		}
		if tCase.isErr == false {
			ExpectEqual("util TranUTCtoLocalTime I", i, t.Errorf, tCase.localString,
				ret)
		}
	}
}

type TestTranTimestamptoLocalTimeType struct {
	timeStamp int64
	timeForm  string
	localTime string
}

func TestTranTimestamptoLocalTime(t *testing.T) {
	testCases := []TestTranTimestamptoLocalTimeType{
		TestTranTimestamptoLocalTimeType{
			timeStamp: 1512478110,
			timeForm:  "2006-01-02 15:04:05",
			localTime: "2017-12-05 20:48:30",
		},
	}
	for i, tCase := range testCases {
		ret := TranTimestamptoLocalTime(tCase.timeStamp, tCase.timeForm)
		ExpectEqual("util TranTimestamptoLocalTime I", i, t.Errorf, tCase.localTime, ret)
	}
}

type FilterSpaceType struct {
	oArray []string
	nArray []string
}

func TestFilterSpace(t *testing.T) {
	testCases := []FilterSpaceType{
		FilterSpaceType{
			oArray: []string{},
			nArray: []string{},
		},
		FilterSpaceType{
			oArray: []string{"", "", ""},
			nArray: []string{},
		},
		FilterSpaceType{
			oArray: []string{"", "32432  ", ""},
			nArray: []string{"32432  "},
		},
		FilterSpaceType{
			oArray: []string{"123", "32432  ", ""},
			nArray: []string{"123", "32432  "},
		},
	}

	for i, tCase := range testCases {
		ret := FilterSpace(tCase.oArray)
		if len(tCase.nArray) != 0 && len(ret) != 0 {
			ExpectEqual("util FilterSpace I", i, t.Errorf, tCase.nArray, ret)
		}
	}
}

type absType struct {
	orgPath string
	absPath string
	isSuc   bool
}

func TestAbs(t *testing.T) {
	thisDir, err := filepath.Abs("./")
	if err != nil {
		t.Errorf("Get current path failed! error: %s", err)
		return
	}
	homeDir, err := GetHomeDirOfUser()
	if err != nil {
		t.Errorf("Get user home dir failed! error: %s", err)
		return
	}
	pathLists := strings.Split(thisDir, OsPathSeparator)

	testCases := []absType{
		absType{
			orgPath: "",
			absPath: thisDir,
			isSuc:   true,
		},
		absType{
			orgPath: "abs",
			absPath: filepath.Join(thisDir, "abs"),
			isSuc:   true,
		},
		absType{
			orgPath: "./abc/enfg/",
			absPath: filepath.Join(thisDir, "abc/enfg/"),
			isSuc:   true,
		},
		absType{
			orgPath: "../abc",
			absPath: filepath.Join(GetParentPath(pathLists, 1), "abc"),
			isSuc:   true,
		},
		absType{
			orgPath: "../../abc",
			absPath: filepath.Join(GetParentPath(pathLists, 2), "abc"),
			isSuc:   true,
		},
		absType{
			orgPath: "../../",
			absPath: GetParentPath(pathLists, 2),
			isSuc:   true,
		},
		absType{
			orgPath: "~/.abd",
			absPath: filepath.Join(homeDir, ".abd"),
			isSuc:   true,
		},
		absType{
			orgPath: "~",
			absPath: homeDir,
			isSuc:   true,
		},
		absType{
			orgPath: "~/",
			absPath: homeDir,
			isSuc:   true,
		},
	}

	for i, tCase := range testCases {
		ret, err := Abs(tCase.orgPath)
		ExpectEqual("util Abs I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if tCase.isSuc && err == nil {
			ExpectEqual("util Abs II", i+1, t.Errorf, tCase.absPath, ret)
		} else if err != nil {
			t.Logf("error: %s", err.Error())
		}
	}
}
