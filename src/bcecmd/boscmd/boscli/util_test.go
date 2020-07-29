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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

import (
	"bceconf"
	"utils/util"
)

var (
	userHomeDir string
)

func init() {
	var err error
	if userHomeDir, err = util.GetHomeDirOfUser(); err != nil {
		fmt.Printf("Error: init get home dir %s\n", err.Error())
		os.Exit(1)
	}
}

func initConfig() error {
	userPath, err := util.GetHomeDirOfUser()
	if err != nil {
		return fmt.Errorf("Get Home of user failed")
	}
	testConfPath := filepath.Join(userPath, bceconf.DEFAULT_FOLDER_IN_USER_HOME)
	bceconf.InitConfig(testConfPath)
	return nil
}

func TestPrintIfNotQuiet(t *testing.T) {
	printIfNotQuiet("%s", 123, "345")
	printIfNotQuiet("%s", 123)
	printIfNotQuiet("%s")
}

type getStorageClassFromStrType struct {
	input  string
	output string
	code   BosCliErrorCode
}

func TestGetStorageClassFromStr(t *testing.T) {
	testCases := []getStorageClassFromStrType{
		getStorageClassFromStrType{
			input:  "",
			output: "",
			code:   BOSCLI_OK,
		},
		getStorageClassFromStrType{
			input:  "STANDARD",
			output: "STANDARD",
			code:   BOSCLI_OK,
		},
		getStorageClassFromStrType{
			input:  "STANDARD_IA",
			output: "STANDARD_IA",
			code:   BOSCLI_OK,
		},
		getStorageClassFromStrType{
			input:  "STANDAIA",
			output: "",
			code:   BOSCLI_UNSUPPORT_STORAGE_CLASS,
		},
		getStorageClassFromStrType{
			input:  "COLD",
			output: "COLD",
			code:   BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, retCode := getStorageClassFromStr(tCase.input)
		util.ExpectEqual("tools.go getStorageClassFromStr I", i+1, t.Errorf, tCase.code, retCode)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("tools.go getStorageClassFromStr II", i+1, t.Errorf, tCase.output, ret)
		}
	}
}

type isTheSameBucketAndObjectType struct {
	srcBucket       string
	srcObject       string
	dstBucket       string
	dstObject       string
	newStorageClass string
	oldStorageClass string
	ret             bool
}

func TestIsTheSameBucketAndObject(t *testing.T) {
	testCases := []isTheSameBucketAndObjectType{
		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123",
			dstBucket:       "123",
			dstObject:       "123",
			newStorageClass: "STANDARD",
			oldStorageClass: "STANDARD",
			ret:             true,
		},
		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123",
			dstBucket:       "123",
			dstObject:       "123",
			newStorageClass: "",
			oldStorageClass: "",
			ret:             false,
		},
		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123",
			dstBucket:       "123",
			dstObject:       "123",
			newStorageClass: "",
			oldStorageClass: "STANDARD",
			ret:             false,
		},

		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123",
			dstBucket:       "123",
			dstObject:       "1",
			newStorageClass: "STANDARD",
			oldStorageClass: "STANDARD",
			ret:             false,
		},
		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123",
			dstBucket:       "123",
			dstObject:       "123",
			newStorageClass: "",
			oldStorageClass: "STANDARD_IA",
			ret:             false,
		},
		isTheSameBucketAndObjectType{
			srcBucket:       "1",
			srcObject:       "123",
			dstBucket:       "123",
			dstObject:       "123",
			newStorageClass: "STANDARD",
			oldStorageClass: "STANDARD",
			ret:             false,
		},
		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123",
			dstBucket:       "123",
			dstObject:       "/123",
			newStorageClass: "STANDARD",
			oldStorageClass: "STANDARD",
			ret:             true,
		},
		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123",
			dstBucket:       "123",
			dstObject:       "/123",
			newStorageClass: "STANDARD",
			oldStorageClass: "STANDARD",
			ret:             true,
		},
		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123",
			dstBucket:       "123",
			dstObject:       "/123",
			newStorageClass: "STANDARD",
			oldStorageClass: "STANDARD",
			ret:             true,
		},
		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123/234",
			dstBucket:       "123",
			dstObject:       "/123//234",
			newStorageClass: "STANDARD",
			oldStorageClass: "STANDARD",
			ret:             true,
		},
		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123/ /234",
			dstBucket:       "123",
			dstObject:       "/123/234",
			newStorageClass: "STANDARD",
			oldStorageClass: "STANDARD",
			ret:             true,
		},
		isTheSameBucketAndObjectType{
			srcBucket:       "123",
			srcObject:       "123/ /234/1/2",
			dstBucket:       "123",
			dstObject:       "/123/234",
			newStorageClass: "STANDARD",
			oldStorageClass: "STANDARD",
			ret:             false,
		},
	}
	for i, tCase := range testCases {
		retCode := isTheSameBucketAndObject(tCase.srcBucket, tCase.srcObject, tCase.dstBucket,
			tCase.dstObject, tCase.newStorageClass, tCase.oldStorageClass)
		util.ExpectEqual("tools.go isTheSameBucketAndObject I", i+1, t.Errorf, tCase.ret, retCode)
	}
}

type splitBosBucketKeyType struct {
	bosPath    string
	bucketName string
	objectKey  string
}

func TestSplitBosBucketKey(t *testing.T) {
	testCases := []splitBosBucketKeyType{
		splitBosBucketKeyType{
			bosPath:    "",
			bucketName: "",
			objectKey:  "",
		},
		splitBosBucketKeyType{
			bosPath:    "bos:/bucket/",
			bucketName: "bucket",
			objectKey:  "",
		},
		splitBosBucketKeyType{
			bosPath:    "bos://bucket/",
			bucketName: "bucket",
			objectKey:  "",
		},
		splitBosBucketKeyType{
			bosPath:    "bucket/",
			bucketName: "bucket",
			objectKey:  "",
		},
		splitBosBucketKeyType{
			bosPath:    "/bucket/object/",
			bucketName: "bucket",
			objectKey:  "object/",
		},
		splitBosBucketKeyType{
			bosPath:    "bucket/object/key",
			bucketName: "bucket",
			objectKey:  "object/key",
		},
		splitBosBucketKeyType{
			bosPath:    "bucket/object/key//.key2",
			bucketName: "bucket",
			objectKey:  "object/key/.key2",
		},
		splitBosBucketKeyType{
			bosPath:    "bucket/object/key/.key2//",
			bucketName: "bucket",
			objectKey:  "object/key/.key2/",
		},
		splitBosBucketKeyType{
			bosPath:    "bos://bucket/object/key/.key2//",
			bucketName: "bucket",
			objectKey:  "object/key/.key2/",
		},
		splitBosBucketKeyType{
			bosPath:    "///",
			bucketName: "",
			objectKey:  "",
		},
	}
	for i, tCase := range testCases {
		bucket, object := splitBosBucketKey(tCase.bosPath)
		util.ExpectEqual("tools.go splitBosBucketKey I", i+1, t.Errorf, tCase.bucketName, bucket)
		util.ExpectEqual("tools.go splitBosBucketKey II", i+1, t.Errorf, tCase.objectKey, object)
	}
}

type calcDstObjectKeyType struct {
	src  string
	dst  string
	ret  string
	code BosCliErrorCode
}

func TestCalcDstObjectKey(t *testing.T) {
	testCases := []calcDstObjectKeyType{
		calcDstObjectKeyType{
			src:  "",
			dst:  "key",
			code: BOSCLI_INTERNAL_ERROR,
		},
		calcDstObjectKeyType{
			src:  "temp",
			dst:  "key",
			ret:  "key",
			code: BOSCLI_OK,
		},
		calcDstObjectKeyType{
			src:  "temp/ew",
			dst:  "",
			ret:  "ew",
			code: BOSCLI_OK,
		},
		calcDstObjectKeyType{
			src:  "temp/ /ew",
			dst:  "test/324/123",
			ret:  "test/324/123",
			code: BOSCLI_OK,
		},
		calcDstObjectKeyType{
			src:  "temp/ /ew",
			dst:  "test/324/",
			ret:  "test/324/ew",
			code: BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, retCode, _ := calcDstObjectKey(tCase.src, tCase.dst)
		util.ExpectEqual("tools.go calcDstObjectKey I", i+1, t.Errorf, tCase.code, retCode)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("tools.go calcDstObjectKey II", i+1, t.Errorf, tCase.ret, ret)
		}
	}
}

type getFinalObjectKeyFromLocalPathType struct {
	src   string
	file  string
	dst   string
	isDir bool
	ret   string
}

var getFinalObjectKeyFromLocalPathCases []getFinalObjectKeyFromLocalPathType

func getCasesOfGetFinalObjectKeyFromLocalPath() {
	if runtime.GOOS == "windows" {
		getFinalObjectKeyFromLocalPathCases = []getFinalObjectKeyFromLocalPathType{
			getFinalObjectKeyFromLocalPathType{
				src:   "D:\\mycode\\debug_bos",
				file:  "D:\\mycode\\debug_bos\\cover.html",
				dst:   "",
				isDir: true,
				ret:   "cover.html",
			},
			getFinalObjectKeyFromLocalPathType{
				src:   "D:\\mycode",
				file:  "D:\\mycode\\debug_bos\\cover.html",
				dst:   "cover.html",
				isDir: false,
				ret:   "cover.html",
			},
			getFinalObjectKeyFromLocalPathType{
				src:   "D:\\mycode",
				file:  "D:\\mycode\\debug_bos\\cover.html",
				dst:   "debug_bos/",
				isDir: true,
				ret:   "debug_bos/debug_bos/cover.html",
			},
			getFinalObjectKeyFromLocalPathType{
				src:   "D:\\mycode\\test",
				file:  "D:\\mycode\\test",
				dst:   "debug_bos/",
				isDir: false,
				ret:   "debug_bos/test",
			},
		}
	} else {
		getFinalObjectKeyFromLocalPathCases = []getFinalObjectKeyFromLocalPathType{
			getFinalObjectKeyFromLocalPathType{
				src:   "/mycode/debug_bos",
				file:  "/mycode/debug_bos/cover.html",
				dst:   "",
				isDir: true,
				ret:   "cover.html",
			},
			getFinalObjectKeyFromLocalPathType{
				src:   "/mycode/",
				file:  "/mycode/debug_bos/cover.html",
				dst:   "cover.html",
				isDir: false,
				ret:   "cover.html",
			},
			getFinalObjectKeyFromLocalPathType{
				src:   "/mycode",
				file:  "/mycode/debug_bos/cover.html",
				dst:   "debug_bos/",
				isDir: true,
				ret:   "debug_bos/debug_bos/cover.html",
			},
			getFinalObjectKeyFromLocalPathType{
				src:   "/mycode/test",
				file:  "/mycode/test",
				dst:   "debug_bos/",
				isDir: false,
				ret:   "debug_bos/test",
			},
			getFinalObjectKeyFromLocalPathType{
				src:   "/mycode/test",
				file:  "/mycode/test",
				dst:   "",
				isDir: false,
				ret:   "test",
			},
		}
	}
}

func TestGetFinalObjectKeyFromLocalPath(t *testing.T) {
	getCasesOfGetFinalObjectKeyFromLocalPath()
	for i, tCase := range getFinalObjectKeyFromLocalPathCases {
		ret := getFinalObjectKeyFromLocalPath(tCase.src, tCase.file, tCase.dst,
			tCase.isDir)
		util.ExpectEqual("tools.go getFinalObjectKeyFromLocalPath II", i+1, t.Errorf, tCase.ret,
			ret)
	}
}

type getObjectNameFromObjectKeyType struct {
	src string
	ret string
}

func TestGetObjectNameFromObjectKey(t *testing.T) {
	testCases := []getObjectNameFromObjectKeyType{
		getObjectNameFromObjectKeyType{
			src: "",
			ret: "",
		},
		getObjectNameFromObjectKeyType{
			src: "temp",
			ret: "temp",
		},
		getObjectNameFromObjectKeyType{
			src: "temp/ew",
			ret: "ew",
		},
		getObjectNameFromObjectKeyType{
			src: "temp/ /ew",
			ret: "ew",
		},
		getObjectNameFromObjectKeyType{
			src: "/ ",
			ret: "",
		},
	}
	for i, tCase := range testCases {
		ret := getObjectNameFromObjectKey(tCase.src)
		if len(tCase.ret) == 0 {
			util.ExpectEqual("tools.go getObjectNameFromObjectKey I", i+1, t.Errorf, 0,
				len(ret))
		} else {
			util.ExpectEqual("tools.go getObjectNameFromObjectKey I", i+1, t.Errorf, tCase.ret,
				ret)
		}
	}
}

type splitPathAndFileType struct {
	src  string
	path string
	file string
}

var splitPathAndFileCases []splitPathAndFileType

func getsplitPathAndFileCases() {
	if runtime.GOOS == "windows" {
		splitPathAndFileCases = []splitPathAndFileType{
			splitPathAndFileType{
				src:  "D:\\mycode\\debug_bos",
				path: "D:\\mycode",
				file: "debug_bos",
			},
			splitPathAndFileType{
				src:  "D:\\mycode\\debug_bos\\cover.html",
				path: "D:\\mycode\\debug_bos",
				file: "cover.html",
			},
			splitPathAndFileType{
				src:  "D:\\mycode\\test",
				path: "D:\\mycode",
				file: "test",
			},
			splitPathAndFileType{
				src:  "D:\\mycode\\test\\",
				path: "D:\\mycode\\test",
				file: "",
			},
			splitPathAndFileType{
				src:  ".\\test\\",
				path: ".\\test",
				file: "",
			},
			splitPathAndFileType{
				src:  ".\\test",
				path: ".",
				file: "test",
			},
		}
	} else {
		splitPathAndFileCases = []splitPathAndFileType{
			splitPathAndFileType{
				src:  "/mycode/debug_bos",
				path: "/mycode",
				file: "debug_bos",
			},
			splitPathAndFileType{
				src:  "/mycode/",
				path: "/mycode",
				file: "",
			},
			splitPathAndFileType{
				src:  "/mycode/debug_bos/cover.html",
				path: "/mycode/debug_bos",
				file: "cover.html",
			},
			splitPathAndFileType{
				src:  "/mycode/test",
				path: "/mycode",
				file: "test",
			},
			splitPathAndFileType{
				src:  "/mycode/test/",
				path: "/mycode/test",
				file: "",
			},
			splitPathAndFileType{
				src:  "./test/",
				path: "./test",
				file: "",
			},
			splitPathAndFileType{
				src:  "./test",
				path: ".",
				file: "test",
			},
		}
	}
}

func TestSplitPathAndFile(t *testing.T) {
	getsplitPathAndFileCases()
	for i, tCase := range splitPathAndFileCases {
		path, file := splitPathAndFile(tCase.src)
		util.ExpectEqual("tools.go splitPathAndFile I", i+1, t.Errorf, tCase.path, path)
		util.ExpectEqual("tools.go splitPathAndFile II", i+1, t.Errorf, tCase.file, file)
	}
}

func initListFileCases(pathPrefix string, needBadLink bool) error {
	paths := []string{
		pathPrefix + "_failed",
		pathPrefix + "/",
		pathPrefix + "/aDir/",
		pathPrefix + "/bDir/",
		pathPrefix + "/cDir/",
		pathPrefix + "/cDir/aDir/bDir/",
		pathPrefix + "/dDir/",
		pathPrefix + "_eDir/",
		pathPrefix + "_fDir/",
	}
	files := []string{
		pathPrefix + "_failed/336",
		pathPrefix + "/ab",
		pathPrefix + "/ac",
		pathPrefix + "/中文",
		pathPrefix + "/aDir/234",
		pathPrefix + "/aDir/345",
		pathPrefix + "/cDir/aDir/bDir/123",
		pathPrefix + "/cDir/aDir/bDir/234",
		pathPrefix + "_eDir/456",
		pathPrefix + "_eDir/567",
		pathPrefix + "_fDir/678",
		pathPrefix + "_fDir/789",
	}
	os.RemoveAll(pathPrefix)
	os.RemoveAll(pathPrefix + "_eDir")
	os.RemoveAll(pathPrefix + "_fDir")
	for _, path := range paths {
		path = strings.Replace(path, "/", util.OsPathSeparator, -1)
		err := util.TryMkdir(path)
		if err != nil {
			return err
		}
	}
	for _, file := range files {
		file = strings.Replace(file, "/", util.OsPathSeparator, -1)
		fd, err := os.Create(file)
		fmt.Fprintf(fd, "%s", file)
		if err != nil {
			return err
		}
		fd.Close()
	}
	// creat symbolic link dir
	symAbsPath, _ := filepath.Abs(pathPrefix + "_eDir")
	if err := os.Symlink(symAbsPath, pathPrefix+"/cDir/eLDir"); err != nil {
		return err
	}
	// creat symbolic link dir
	symAbsPath, _ = filepath.Abs(pathPrefix + "/cDir")
	if err := os.Symlink(symAbsPath, pathPrefix+"/aDir/cDir"); err != nil {
		return err
	}
	// creat symbolic link dir
	symAbsPath, _ = filepath.Abs(pathPrefix + "_fDir")
	if err := os.Symlink(symAbsPath, pathPrefix+"/aDir/fDir"); err != nil {
		return err
	}

	// creat symbolic link file
	symAbsPath, _ = filepath.Abs(pathPrefix + "/cDir/aDir/bDir/123")
	if err := os.Symlink(symAbsPath, pathPrefix+"/aDir/235"); err != nil {
		return err
	}
	// creat bad symbolic link file
	symAbsPath, _ = filepath.Abs(pathPrefix + "_failed/336")
	if err := os.Symlink(symAbsPath, pathPrefix+"/337"); err != nil {
		return err
	}
	if needBadLink {
		return os.RemoveAll(pathPrefix + "_failed")
	}
	return nil
}

func removeListFileCases(pathPrefix string) error {
	if err := os.RemoveAll(pathPrefix); err != nil {
		return fmt.Errorf("remove %s failed", pathPrefix)
	}
	if err := os.RemoveAll(pathPrefix + "_eDir"); err != nil {
		return fmt.Errorf("remove %s_eDir failed", pathPrefix)
	}
	if err := os.RemoveAll(pathPrefix + "_fDir"); err != nil {
		return fmt.Errorf("remove %s_fDir failed", pathPrefix)
	}
	os.RemoveAll(pathPrefix + "_failed")
	return nil
}

type listLocalFileType struct {
	path           string
	fileList       []string
	haveFileErr    bool
	isSuc          bool
	followSymlinks bool

	// filter
	exclude     []string
	include     []string
	excludeTime []string
	includeTime []string
}

func TestListLocalFiles(t *testing.T) {
	pathPrefix := "test_list_local"
	if err := initListFileCases(pathPrefix, true); err != nil {
		t.Errorf("tools.go listLocal initListFileCases failed: %s", err.Error())
		return
	}
	absPath, err := filepath.Abs(pathPrefix)
	if err != nil {
		t.Errorf("tools.go listLocal get abs path failed: %s", err.Error())
		return
	}
	badAbsPath, err := filepath.Abs(pathPrefix + "_failed")
	if err != nil {
		t.Errorf("tools.go listLocal get abs path failed: %s", err.Error())
		return
	}
	if strings.HasSuffix(absPath, util.OsPathSeparator) {
		t.Errorf("tools.go listLocal absolute path end with / or \\")
		return
	}
	listLocalFileCases := []listLocalFileType{
		//1
		listLocalFileType{
			path: absPath,
			fileList: []string{"aDir/234", "aDir/235", "aDir/345", "aDir/cDir/aDir/bDir/123",
				"aDir/cDir/aDir/bDir/234", "aDir/cDir/eLDir/456", "aDir/cDir/eLDir/567",
				"aDir/fDir/678", "aDir/fDir/789", "ab", "ac", "cDir/aDir/bDir/123",
				"cDir/aDir/bDir/234", "cDir/eLDir/456", "cDir/eLDir/567", "中文"},
			isSuc:          true,
			haveFileErr:    true,
			followSymlinks: true,
		},
		listLocalFileType{
			path:           absPath + "/ab",
			fileList:       []string{"ab"},
			isSuc:          true,
			followSymlinks: true,
		},
		//3
		listLocalFileType{
			path:           absPath + "/bDir/",
			fileList:       []string{},
			isSuc:          true,
			followSymlinks: true,
		},
		listLocalFileType{
			path: absPath + "/aDir/",
			fileList: []string{"234", "235", "345", "cDir/aDir/bDir/123", "cDir/aDir/bDir/234",
				"cDir/eLDir/456", "cDir/eLDir/567", "fDir/678", "fDir/789"},
			isSuc:          true,
			followSymlinks: true,
		},
		//5
		listLocalFileType{
			path:           "/root/",
			isSuc:          false,
			followSymlinks: true,
		},
		listLocalFileType{
			path:           badAbsPath + "/root/",
			isSuc:          true,
			haveFileErr:    true,
			fileList:       []string{},
			followSymlinks: true,
		},
		// 7
		listLocalFileType{
			path:           badAbsPath + "/336",
			isSuc:          true,
			haveFileErr:    true,
			fileList:       []string{},
			followSymlinks: true,
		},
		listLocalFileType{
			path: absPath,
			fileList: []string{"aDir/234", "aDir/345", "aDir/cDir/aDir/bDir/123",
				"aDir/cDir/aDir/bDir/234", "aDir/cDir/eLDir/456", "aDir/cDir/eLDir/567",
				"aDir/fDir/678", "aDir/fDir/789", "ab", "ac", "cDir/aDir/bDir/123",
				"cDir/aDir/bDir/234", "cDir/eLDir/456", "cDir/eLDir/567", "中文"},
			haveFileErr:    true,
			isSuc:          true,
			followSymlinks: false,
		},
		//9 filter: exclude
		listLocalFileType{
			path:           absPath,
			exclude:        []string{"./test_list_local/*"},
			fileList:       []string{},
			isSuc:          true,
			haveFileErr:    true,
			followSymlinks: true,
		},
		//10 filter: exclude
		listLocalFileType{
			path:    absPath,
			exclude: []string{"./test_list_local/aDir/*"},
			fileList: []string{"ab", "ac", "cDir/aDir/bDir/123",
				"cDir/aDir/bDir/234", "cDir/eLDir/456", "cDir/eLDir/567", "中文"},
			isSuc:          true,
			haveFileErr:    true,
			followSymlinks: true,
		},
		//11 filter: exclude
		listLocalFileType{
			path: absPath,
			exclude: []string{"./test_list_local/aDir/*",
				"./test_list_local/cDir/*"},
			fileList:       []string{"ab", "ac", "中文"},
			isSuc:          true,
			haveFileErr:    true,
			followSymlinks: true,
		},
		//12 filter: include
		listLocalFileType{
			path:    absPath,
			include: []string{"./test_list_local/*"},
			fileList: []string{"aDir/234", "aDir/235", "aDir/345", "aDir/cDir/aDir/bDir/123",
				"aDir/cDir/aDir/bDir/234", "aDir/cDir/eLDir/456", "aDir/cDir/eLDir/567",
				"aDir/fDir/678", "aDir/fDir/789", "ab", "ac", "cDir/aDir/bDir/123",
				"cDir/aDir/bDir/234", "cDir/eLDir/456", "cDir/eLDir/567", "中文"},
			isSuc:          true,
			haveFileErr:    true,
			followSymlinks: true,
		},
		//13 filter: include abs path
		listLocalFileType{
			path:    absPath,
			include: []string{absPath + util.OsPathSeparator + "*"},
			fileList: []string{"aDir/234", "aDir/235", "aDir/345", "aDir/cDir/aDir/bDir/123",
				"aDir/cDir/aDir/bDir/234", "aDir/cDir/eLDir/456", "aDir/cDir/eLDir/567",
				"aDir/fDir/678", "aDir/fDir/789", "ab", "ac", "cDir/aDir/bDir/123",
				"cDir/aDir/bDir/234", "cDir/eLDir/456", "cDir/eLDir/567", "中文"},
			isSuc:          true,
			haveFileErr:    true,
			followSymlinks: true,
		},
		//14 filter: muilt include
		listLocalFileType{
			path:           absPath,
			include:        []string{"./test_list_local/aDir/234", "./test_list_local/中文"},
			fileList:       []string{"aDir/234", "中文"},
			isSuc:          true,
			haveFileErr:    true,
			followSymlinks: true,
		},
		//15
		listLocalFileType{
			path:           absPath,
			include:        []string{"./test_list_local/cDir/*/234"},
			fileList:       []string{"cDir/aDir/bDir/234"},
			isSuc:          true,
			haveFileErr:    true,
			followSymlinks: true,
		},
	}

	for i, tCase := range listLocalFileCases {
		var (
			filter   *bosFilter = nil
			fRetCode BosCliErrorCode
			fErr     error
		)
		fmt.Println("\nstart id:", i+1)
		if len(tCase.exclude) > 0 || len(tCase.include) > 0 || len(tCase.includeTime) > 0 ||
			len(tCase.excludeTime) > 0 {
			filter, fRetCode, fErr = newSyncFilter(tCase.exclude, tCase.include, tCase.excludeTime,
				tCase.includeTime, true)
			if fRetCode != BOSCLI_OK {
				t.Errorf("ID: %d, get new filter error : %s, code: %s", i+1, fErr, fRetCode)
			}
		}

		list := NewLocalFileIterator(tCase.path, filter, tCase.followSymlinks)
		fileNum := 0
		haveFileErr := false
		if tCase.isSuc {
			for {
				fileInfo, err := list.next()
				if err != nil {
					util.ExpectEqual("tools.go listLocal I", i+1, t.Errorf, true, false)
					break
				}
				if fileInfo.err != nil {
					break
				}
				if fileInfo.ended {
					util.ExpectEqual("tools.go listLocal III", i+1, t.Errorf, len(tCase.fileList),
						fileNum)
					break
				}
				if fileInfo.file.err != nil {
					haveFileErr = true
					continue
				}
				t.Logf("index: %d %s", fileNum, fileInfo.file.key)
				okey := strings.Replace(tCase.fileList[fileNum], "/", util.OsPathSeparator, -1)
				util.ExpectEqual("tools.go listLocal IV", i+1, t.Errorf, okey, fileInfo.file.key)
				util.ExpectEqual("tools.go listLocal V", i+1, t.Errorf, true,
					fileInfo.file.size > 0)
				util.ExpectEqual("tools.go listLocal VI", i+1, t.Errorf, true,
					fileInfo.file.mtime > 0)
				tmepKey := strings.Replace(fileInfo.file.key, "/", util.OsPathSeparator, -1)
				util.ExpectEqual("tools.go listLocal VII", i+1, t.Errorf, true,
					strings.HasSuffix(fileInfo.file.path, tmepKey))
				util.ExpectEqual("tools.go listLocal VIII", i+1, t.Errorf, tCase.haveFileErr,
					haveFileErr)
				fileNum++
			}
		} else {
			haveErr := false
			for {
				fileInfo, err := list.next()
				if err != nil {
					haveErr = true
					break
				}
				if fileInfo.err != nil {
					haveErr = true
					break
				}
				if fileInfo.ended {
					break
				}
				if fileInfo.file.err != nil {
					haveErr = true
				}
			}
			util.ExpectEqual("tools.go listLocal suc I", i+1, t.Errorf, true, haveErr)
		}
	}
	if err := removeListFileCases(pathPrefix); err != nil {
		t.Errorf("tools.go listLocal %s", err.Error())
	}

}

type checkBosPathType struct {
	src string
	ret BosCliErrorCode
}

func TestCheckBosPath(t *testing.T) {
	testCases := []checkBosPathType{
		checkBosPathType{
			src: "",
			ret: BOSCLI_OK,
		},
		checkBosPathType{
			src: "temp",
			ret: BOSCLI_OK,
		},
		checkBosPathType{
			src: "temp/ew",
			ret: BOSCLI_OK,
		},
		checkBosPathType{
			src: "temp/ /ew",
			ret: BOSCLI_OK,
		},
		checkBosPathType{
			src: "/ ",
			ret: BOSCLI_BOSPATH_IS_INVALID,
		},
		checkBosPathType{
			src: "bos:/",
			ret: BOSCLI_OK,
		},
		checkBosPathType{
			src: "bos://",
			ret: BOSCLI_OK,
		},
		checkBosPathType{
			src: "bos://liup",
			ret: BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, _ := checkBosPath(tCase.src)
		util.ExpectEqual("tools.go checkBosPathType I", i+1, t.Errorf, tCase.ret, ret)
	}
}

type trimTrailingSlashType struct {
	src string
	ret string
}

func TestTrimTrailingSlash(t *testing.T) {
	testCases := []trimTrailingSlashType{
		trimTrailingSlashType{
			src: "",
			ret: "",
		},
		trimTrailingSlashType{
			src: "temp",
			ret: "temp",
		},
		trimTrailingSlashType{
			src: "temp/ew",
			ret: "temp/ew",
		},
		trimTrailingSlashType{
			src: "temp/ /ew",
			ret: "temp/ /ew",
		},
		trimTrailingSlashType{
			src: "/ ",
			ret: "",
		},
		trimTrailingSlashType{
			src: "bos:/",
			ret: "bos:",
		},
		trimTrailingSlashType{
			src: "bos://liup//",
			ret: "bos://liup",
		},
	}
	for i, tCase := range testCases {
		ret := trimTrailingSlash(tCase.src)
		util.ExpectEqual("tools.go trimTrailingSlashType I", i+1, t.Errorf, tCase.ret, ret)
	}
}

type replaceToOsPathType struct {
	src string
	ret string
}

func TestReplaceToOsPath(t *testing.T) {
	testCases := []replaceToOsPathType{
		replaceToOsPathType{
			src: "",
			ret: "",
		},
		replaceToOsPathType{
			src: "temp/ew",
			ret: "temp/ew",
		},
		replaceToOsPathType{
			src: "temp/ /ew",
			ret: "temp/ /ew",
		},
		replaceToOsPathType{
			src: "/ ",
			ret: "/ ",
		},
		replaceToOsPathType{
			src: "bos:/",
			ret: "bos:/",
		},
		replaceToOsPathType{
			src: "bos://liup//",
			ret: "bos://liup//",
		},
	}
	for i, tCase := range testCases {
		eRet := strings.Replace(tCase.ret, "/", util.OsPathSeparator, -1)
		ret := replaceToOsPath(tCase.src)
		util.ExpectEqual("tools.go replaceToOsPathType I", i+1, t.Errorf, eRet, ret)
	}
}
