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
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	BOS_PATH_SEPARATOR = "/"
)

var (
	OsPathSeparator = fmt.Sprintf("%c", os.PathSeparator)
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Generate a random string
func GetRandomString(size int64) string {
	var (
		basicString = []rune("123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
		i           int64
	)
	retTemp := make([]rune, size)
	basicLen := len(basicString)
	for i = 0; i < size; i++ {
		retTemp[i] = basicString[rand.Intn(basicLen)]
	}
	return string(retTemp)
}

// Get host name from url
func GetHostFromUrl(useUrl string) (string, error) {
	if useUrl == "" {
		return "", fmt.Errorf("url is empty")
	}
	u, err := url.Parse(useUrl)
	if err != nil {
		return "", err
	}
	hostName := u.Hostname()
	if hostName == "" {
		return "", fmt.Errorf("can't get host name of %s, url must start with http:// or https://", useUrl)
	}
	return hostName, nil
}

// Provide dynamic waiting propmt.
// stop: send signal to stop PrintWaiting
func PrintWaiting(propmt string, stop chan bool) {
	for {
		select {
		case <-stop:
			fmt.Printf("\r%s ", propmt)
			return
		default:
			fmt.Printf("\r%s /", propmt)
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("\r%s \\", propmt)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Check whether path exist
func DoesPathExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	return false
}

// Check whether file path exist.
func DoesFileExist(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err == nil && !fileInfo.IsDir() {
		return true
	}
	return false
}

// Check whether directory exist.
func DoesDirExist(fileFolder string) bool {
	fileInfo, err := os.Stat(fileFolder)
	if err == nil && fileInfo.IsDir() {
		return true
	}
	return false
}

// check dir is writable
// TODO: have more efficient method?
func IsDirWritable(fileFolder string) bool {
	if !DoesDirExist(fileFolder) {
		return false
	}
	tempFilePath := fileFolder
	if !strings.HasSuffix(fileFolder, OsPathSeparator) {
		tempFilePath += OsPathSeparator
	}
	tempFilePath += GetRandomString(20)
	fp, err := os.OpenFile(tempFilePath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return false
	}
	fp.Close()
	os.Remove(tempFilePath)
	return true
}

// check file is writable
func IsFileWritable(filePath string) bool {
	if !DoesFileExist(filePath) {
		return false
	}
	fp, err := os.OpenFile(filePath, os.O_APPEND, os.ModePerm)
	if err != nil {
		return false
	}
	fp.Close()
	return true

}

// check whether file is symbolic link
func IsSymbolicLink(filePath string) (bool, error) {
	ret, err := os.Lstat(filePath)
	if err != nil {
		return false, err
	}
	return (ret.Mode() & os.ModeSymlink) != 0, nil
}

// try to mkdir
func TryMkdir(pathToMake string) error {
	if !DoesPathExist(pathToMake) {
		err := os.MkdirAll(pathToMake, os.ModePerm)
		return err
	} else if DoesFileExist(pathToMake) {
		return fmt.Errorf("can not mkdir %s for it has exist a file with same name", pathToMake)
	}
	return nil
}

// Geting the size of file.
func GetSizeOfFile(filePath string) (int64, error) {
	info, err := os.Lstat(filePath)
	if err != nil {
		return 0, err
	}
	if info.IsDir() {
		return 0, fmt.Errorf("Can't get size of directory")
	}
	return info.Size(), nil //bytes
}

func sortFileNamsWithBosPathSeparator(localPath string, fileOrDirNames []string) {
	// For compare with object in bos, we need change OsPathSeparator to BOS_PATH_SEPARATOR
	if OsPathSeparator != BOS_PATH_SEPARATOR {
		for i, val := range fileOrDirNames {
			fileOrDirNames[i] = strings.Replace(val, OsPathSeparator, BOS_PATH_SEPARATOR, -1)
		}
	}
	// sort
	sort.Strings(fileOrDirNames)

	// hava dir is the prefix of other files or dir?
	// In old code, we check every file path is File or DIR, but it is inefficiency in nfs.
	// Now, we only check the file path which is prefix of other path files
	fileNum := len(fileOrDirNames)
	if fileNum > 1 {
		hasDirIsPrefix := false
		for i := 1; i < fileNum; i++ {
			if strings.HasPrefix(fileOrDirNames[i], fileOrDirNames[i-1]) {
				tempLocaPath := filepath.Join(localPath, fileOrDirNames[i-1])
				if DoesDirExist(tempLocaPath) {
					fileOrDirNames[i-1] += BOS_PATH_SEPARATOR
					hasDirIsPrefix = true
				}
			}
		}
		if hasDirIsPrefix {
			sort.Strings(fileOrDirNames)
		}
	}

	if OsPathSeparator != BOS_PATH_SEPARATOR {
		for i, val := range fileOrDirNames {
			fileOrDirNames[i] = strings.Replace(val, BOS_PATH_SEPARATOR, OsPathSeparator, -1)
		}
	}
}

// Get sorted files and dir names from dir
func ReadSortedDirNames(localPath string) ([]string, error) {
	fp, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	fileOrDirNames, err := fp.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	sortFileNamsWithBosPathSeparator(localPath, fileOrDirNames)

	return fileOrDirNames, nil
}

// Geting home directory of current user
func GetHomeDirOfUser() (string, error) {
	// TODO maybe can't cross platform
	user, err := user.Current()
	if err == nil {
		return user.HomeDir, nil
	}
	return "", err
}

// Wrapper of function filepath.Abs()
// Abs can recognize '~'
func Abs(localPath string) (string, error) {
	if strings.HasPrefix(localPath, "~") {
		userHomeDir, err := GetHomeDirOfUser()
		if err != nil {
			return "", err
		}
		return filepath.Join(userHomeDir, localPath[1:]), nil
	}
	return filepath.Abs(localPath)
}

func PromptConfirm(format string, args ...interface{}) bool {
	fmt.Printf(format+" (Y/N)", args...)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		confirm := strings.TrimSpace(scanner.Text())
		switch confirm {
		case "y", "Y", "yes", "Yes", "YES":
			return true
		case "n", "N", "no", "No", "NO":
			return false
		default:
			fmt.Printf("Only accept Y or N, please confirm:")
		}
	}
	return false
}

// Change UTC time string to timestamp
func TranUTCTimeStringToTimeStamp(utcTimeString, oldTimeForm string) (int64, error) {
	utcTime, err := time.Parse(oldTimeForm, utcTimeString)
	if err != nil {
		return 0, err
	}
	timestamp := utcTime.Unix()
	return timestamp, nil
}

// Change UTC time string to local time string
func TranUTCtoLocalTime(utcTimeString, oldTimeForm, newTimeForm string) (string, error) {
	utcTime, err := time.Parse(oldTimeForm, utcTimeString)
	if err != nil {
		return "", err
	}
	timestamp := utcTime.Unix()
	_, offset := utcTime.Zone()
	localTime := time.Unix(timestamp+int64(offset), 0)
	return localTime.Format(newTimeForm), nil
}

// Change timestamp to local time
func TranTimestamptoLocalTime(timestamp int64, newTimeForm string) string {
	localTime := time.Unix(timestamp, 0)
	return localTime.Format(newTimeForm)
}

// Fileter array
func FilterSpace(components []string) []string {
	var ret []string
	for _, val := range components {
		if strings.TrimSpace(val) != "" {
			ret = append(ret, val)
		}
	}
	return ret
}

// Get md5 of string
func StringMd5(strVal string) string {
	md5New := md5.New()
	io.WriteString(md5New, strVal)
	return hex.EncodeToString(md5New.Sum(nil))
}

// calc the md5 of local file
// whence: 0 means relative to the origin of the file, 1 means relative to the current offset,
// and 2 means relative to the end
func GetFileMd5(fd *os.File, offset, size int64, whence int) (string, error) {
	if fd == nil {
		return "", fmt.Errorf("fd is null!")
	}
	if _, err := fd.Seek(offset, whence); err != nil {
		return "", err
	}
	md5New := md5.New()
	if copied, err := io.CopyN(md5New, fd, size); err != nil {
		return "", err
	} else if copied != size {
		return "", fmt.Errorf("can not copy %d bytes when offset is %d and whence is %d",
			size, offset, whence)
	}
	return hex.EncodeToString(md5New.Sum(nil)), nil
}
