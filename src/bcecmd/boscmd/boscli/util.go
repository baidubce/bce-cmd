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
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

import (
	"bcecmd/boscmd"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
	"utils/util"
)

// Print msg if not quiet
func printIfNotQuiet(format string, args ...interface{}) (int, error) {
	if !Quiet {
		return fmt.Printf(format, args...)
	}
	return 0, nil
}

// Check whether storage class is correct
func getStorageClassFromStr(str string) (string, BosCliErrorCode) {
	switch strings.ToUpper(str) {
	case "":
		return "", BOSCLI_OK
	case api.STORAGE_CLASS_STANDARD:
		return api.STORAGE_CLASS_STANDARD, BOSCLI_OK
	case api.STORAGE_CLASS_STANDARD_IA:
		return api.STORAGE_CLASS_STANDARD_IA, BOSCLI_OK
	case api.STORAGE_CLASS_COLD:
		return api.STORAGE_CLASS_COLD, BOSCLI_OK
	}
	return "", BOSCLI_UNSUPPORT_STORAGE_CLASS
}

// Check whether two bospath is the same
func isTheSameBucketAndObject(srcBucket, srcObject, dstBucket, dstObject, newStorageClass,
	oldStorageClass string) bool {
	if srcBucket == dstBucket && isSameObjectName(srcObject, dstObject) {
		if oldStorageClass == "" {
			oldStorageClass = DEFAULT_STORAGE_CLASS
		}
		if oldStorageClass == newStorageClass {
			return true
		}
	}
	return false
}

// Split bos path to bucket name and object key
func splitBosBucketKey(bosPath string) (string, string) {
	if bosPath == "" {
		return "", ""
	}

	if strings.HasPrefix(bosPath, BOS_PATH_PREFIX_DOUBLE) {
		return extractBucketNameAndKey(bosPath[len(BOS_PATH_PREFIX_DOUBLE):])
	}

	if strings.HasPrefix(bosPath, BOS_PATH_PREFIX) {
		return extractBucketNameAndKey(bosPath[len(BOS_PATH_PREFIX):])
	}
	return extractBucketNameAndKey(bosPath)
}

func extractBucketNameAndKey(bosPath string) (string, string) {
	var (
		bosKey     string
		bucketName string
	)
	bosPath = strings.TrimSpace(bosPath)
	bosComponents := util.FilterSpace(strings.Split(bosPath, boscmd.BOS_PATH_SEPARATOR))
	bosComponentsLen := len(bosComponents)
	if bosComponentsLen == 0 {
		return "", ""
	}
	bucketName = bosComponents[0]
	if bosComponentsLen > 1 {
		bosKey = strings.Join(bosComponents[1:], boscmd.BOS_PATH_SEPARATOR)
		if strings.HasSuffix(bosPath, boscmd.BOS_PATH_SEPARATOR) && !strings.HasSuffix(bosKey,
			boscmd.BOS_PATH_SEPARATOR) {
			bosKey += boscmd.BOS_PATH_SEPARATOR
		}
	}
	return bucketName, bosKey
}

// Check whether two object keys is the same
func isSameObjectName(srcObjectKey, dstObjectKey string) bool {
	var (
		srcIndex int
		dstIndex int
	)

	srcComponents := util.FilterSpace(strings.Split(srcObjectKey, boscmd.BOS_PATH_SEPARATOR))
	dstComponents := util.FilterSpace(strings.Split(dstObjectKey, boscmd.BOS_PATH_SEPARATOR))
	lenSrcComponents := len(srcComponents)
	lenDstComponents := len(dstComponents)

	// lenDstComponents must equal to lenSrcComponent or one more than lenSrcComponent
	absVal := lenSrcComponents - lenDstComponents

	if absVal != 1 && absVal != 0 {
		return false
	}

	for {
		if srcIndex < lenSrcComponents && dstIndex < lenDstComponents {
			if dstComponents[dstIndex] != srcComponents[srcIndex] {
				return false
			}
			srcIndex += 1
			dstIndex += 1
		} else {
			break
		}
	}
	return true
}

func calcDstObjectKey(srcObjectKey, dstObjectKey string) (string, BosCliErrorCode, error) {
	var tempObject string
	index := strings.LastIndex(srcObjectKey, boscmd.BOS_PATH_SEPARATOR)
	if index == -1 {
		tempObject = srcObjectKey
	} else {
		tempObject = srcObjectKey[index+1:]
	}
	if tempObject == "" {
		return "", BOSCLI_INTERNAL_ERROR, fmt.Errorf("Source object name error %s", srcObjectKey)
	}
	if dstObjectKey == "" {
		return tempObject, BOSCLI_OK, nil
	}
	if strings.HasSuffix(dstObjectKey, boscmd.BOS_PATH_SEPARATOR) {
		return dstObjectKey + tempObject, BOSCLI_OK, nil
	}
	return dstObjectKey, BOSCLI_OK, nil
}

// get final object key from local file path
func getFinalObjectKeyFromLocalPath(srcPath, filePath, dstObjectKey string, srcIsDir bool) string {

	if srcIsDir {
		relativePath := filePath[len(srcPath)+1:]
		relativePath = strings.Replace(relativePath, util.OsPathSeparator,
			boscmd.BOS_PATH_SEPARATOR, -1)
		return dstObjectKey + relativePath
	}

	finalObjectKey := filepath.Base(filePath)
	if dstObjectKey == "" {
		return finalObjectKey
	} else if strings.HasSuffix(dstObjectKey, boscmd.BOS_PATH_SEPARATOR) {
		return dstObjectKey + finalObjectKey
	} else {
		return dstObjectKey
	}
}

// Get object name from objet key
func getObjectNameFromObjectKey(objectKey string) string {
	srcComponents := strings.Split(strings.TrimSpace(objectKey), boscmd.BOS_PATH_SEPARATOR)
	lenSrcComponents := len(srcComponents)
	if lenSrcComponents == 0 {
		return ""
	}
	return srcComponents[lenSrcComponents-1]
}

// Split local path into dir and file name
// path: not end with OsPathSeparator
// file: not start with OsPathSeparator
func splitPathAndFile(inputPath string) (string, string) {
	inputPath = strings.TrimSpace(inputPath)
	components := strings.Split(inputPath, util.OsPathSeparator)
	lenComponents := len(components)

	retFile := components[lenComponents-1]
	retPath := ""

	if lenComponents > 1 {
		pathComponents := util.FilterSpace(components[:lenComponents-1])
		if len(pathComponents) > 0 {
			retPath = strings.Join(pathComponents, util.OsPathSeparator)
		}
	}
	if strings.HasPrefix(inputPath, util.OsPathSeparator) {
		retPath = util.OsPathSeparator + retPath
	}
	return retPath, retFile
}

// list local directory (only list files)
// localPath must be absolute path
// file name is sorted by lexicographical ordeor
// NOTICE: when sort file names, os separator is replace to bos separator
func NewLocalFileIterator(localPath string, filter *bosFilter,
	followSymlinks bool) *LocalFileIterator {

	localPathPrefix := ""
	if util.DoesDirExist(localPath) {
		localPathPrefix = localPath
	} else if util.DoesFileExist(localPath) {
		localPathPrefix = filepath.Dir(localPath)
	}

	files := &LocalFileIterator{
		filter:             filter,
		filesChan:          make(chan listFileResult, 100),
		followSymlinks:     followSymlinks,
		localPathPrefixLen: len(localPathPrefix),
	}
	go files.localWalk(localPath)
	return files
}

type LocalFileIterator struct {
	filter             *bosFilter
	filesChan          chan listFileResult
	followSymlinks     bool
	localPathPrefixLen int
}

func (l *LocalFileIterator) localWalk(localPath string) {

	if !util.DoesPathExist(localPath) {
		l.filesChan <- listFileResult{ended: true}
		return
	}

	fileInfo, err := os.Lstat(localPath)
	if err != nil {
		l.filesChan <- listFileResult{err: err}
		l.filesChan <- listFileResult{ended: true}
		return
	}

	l.listAllFiles(localPath, fileInfo)

	// have get all files
	l.filesChan <- listFileResult{ended: true}
}

// lInfo is got by Lstat (don't follow symbolic link)
func (l *LocalFileIterator) listAllFiles(localPath string, lInfo os.FileInfo) {
	isSymbol := ((lInfo.Mode() & os.ModeSymlink) != 0)

	info, err := os.Stat(localPath)
	if err != nil {
		l.filesChan <- listFileResult{
			file: &fileDetail{
				path: localPath,
				err:  err,
			},
		}
		return
	}

	if info.IsDir() {
		if !strings.HasSuffix(localPath, util.OsPathSeparator) {
			localPath += util.OsPathSeparator
		}
		// Pattern Filter, don't apply include filter for directory.
		// because the DIR may don't match the pattern, but it's children matched.
		// e.g. DIR "./test/" don't match with "./test/*.html", howerver, it's child
		// "./test/tet.html" matched.
		if l.filter != nil {
			if filtered, err := l.filter.ExcludePatternFilter(localPath); filtered || err != nil {
				// only error is ErrBadPattern, when it occurs we need to stop list local file
				if err != nil {
					l.filesChan <- listFileResult{err: err}
				}
				return
			}
		}

		// when sort file names, os separator is replace to bos separator
		names, err := util.ReadSortedDirNames(localPath)
		if err != nil {
			l.filesChan <- listFileResult{
				file: &fileDetail{
					path: localPath,
					err:  err,
				},
			}
			return
		}
		for _, name := range names {
			fileName := filepath.Join(localPath, name)
			fileInfo, err := os.Lstat(fileName)
			if err != nil {
				l.filesChan <- listFileResult{
					file: &fileDetail{
						path: fileName,
						err:  err,
					},
				}
			} else {
				l.listAllFiles(fileName, fileInfo)
			}
		}
	} else {
		//check whether file is symbolic link
		if isSymbol && !l.followSymlinks {
			return
		}

		result := listFileResult{
			file: &fileDetail{
				path:     localPath,
				key:      replaceToBosPath(localPath[l.localPathPrefixLen:]),
				realPath: localPath,
				size:     info.Size(),
				mtime:    info.ModTime().Unix(),
				gtime:    time.Now().Unix(),
			},
		}
		// Filter file by patterns
		if l.filter != nil {
			if filtered, err := l.filter.PatternFilter(localPath); filtered || err != nil {
				if err != nil {
					l.filesChan <- listFileResult{err: err}
				}
				return
			}
		}
		// Filter file by time
		if l.filter != nil && l.filter.TimeFilter(result.file.mtime) {
			return
		}
		l.filesChan <- result
	}
}

//Get mtime and size of symbolic link file
func (l *LocalFileIterator) getRelSizeAndMtimeOfLink(localPath string) (*fileDetail, error) {

	relSrcPath, err := filepath.EvalSymlinks(localPath)
	if err != nil {
		return nil, err
	}
	relIinfo, err := os.Stat(relSrcPath)
	if err != nil {
		return nil, err
	}

	return &fileDetail{
		path:     localPath,
		key:      replaceToBosPath(localPath[l.localPathPrefixLen:]),
		realPath: relSrcPath,
		size:     relIinfo.Size(),
		mtime:    relIinfo.ModTime().Unix(),
		gtime:    time.Now().Unix(),
	}, nil
}

// Get next file
// RETURN:
//     1. listFileResult.err != nil     ： serious error occurred, can't continue.
//     2. listFileResult.file.err != nil: error occurred when get info of some file, can continue
//     3. listFileResult.ended == true    : have listed all files
func (l *LocalFileIterator) next() (*listFileResult, error) {
	select {
	case fileDetailVal := <-l.filesChan:
		return &fileDetailVal, nil
	case <-time.After(time.Duration(GET_NET_LOCAL_FILE_TIME_OUT) * time.Millisecond):
		return nil, fmt.Errorf("Get files list time out!")
	}
	return nil, fmt.Errorf("List local files failed")
}

// Check whether bos path is invaild
func checkBosPath(bosPath string) (BosCliErrorCode, error) {
	if bosPath == "" {
		return BOSCLI_OK, nil
	}
	if !strings.HasPrefix(bosPath, BOS_PATH_PREFIX) {
		if strings.HasPrefix(bosPath, boscmd.BOS_PATH_SEPARATOR) {
			return BOSCLI_BOSPATH_IS_INVALID, fmt.Errorf("Invaild BOS path: %s, BOS path must "+
				"start with bos:/ or bos://", bosPath)
		}
	}
	return BOSCLI_OK, nil
}

// Trims all trailing slashes of the give path
func trimTrailingSlash(inputPath string) string {
	inputPath = strings.TrimSpace(inputPath)
	if !strings.HasSuffix(inputPath, boscmd.BOS_PATH_SEPARATOR) {
		return inputPath
	}

	i := len(inputPath) - 1
	for ; i >= 0; i-- {
		if inputPath[i] != '/' {
			break
		}
	}

	return inputPath[:i+1]
}

// Replace system path separator to bos bos separator
func replaceToBosPath(inputPath string) string {
	inputPath = strings.Replace(inputPath, util.OsPathSeparator, boscmd.BOS_PATH_SEPARATOR, -1)
	if strings.HasPrefix(inputPath, boscmd.BOS_PATH_SEPARATOR) {
		inputPath = inputPath[1:]
	}
	return inputPath
}

// Replace bos bos separator and "\\" to system path separator
func replaceToOsPath(inputPath string) string {
	inputPath = strings.Replace(inputPath, "\\", util.OsPathSeparator, -1)
	inputPath = strings.Replace(inputPath, boscmd.BOS_PATH_SEPARATOR, util.OsPathSeparator, -1)
	return inputPath
}

// check whether the error is because no such file or directory
func ErrIsNotExist(err error) bool {
	if os.IsNotExist(err) {
		return true
	}
	return false
}

// remove prefix of bosPath
func FilterPrefixOfBosPath(bosPath string) string {
	if strings.HasPrefix(bosPath, BOS_PATH_PREFIX_DOUBLE) {
		return bosPath[len(BOS_PATH_PREFIX_DOUBLE):]
	} else if strings.HasPrefix(bosPath, BOS_PATH_PREFIX) {
		return bosPath[len(BOS_PATH_PREFIX):]
	}
	return bosPath
}

// get the info of local file
func getFileMate(localPath string) (*fileDetail, error) {
	fileInfo, err := os.Lstat(localPath)
	if err != nil {
		return nil, err
	}

	return &fileDetail{
		path:     localPath,
		realPath: localPath,
		size:     fileInfo.Size(),
		mtime:    fileInfo.ModTime().Unix(),
		gtime:    time.Now().Unix(),
		isDir:    fileInfo.IsDir(),
	}, nil
}

func getCrc32OfLocalFile(localPath string) (string, error) {
	fd, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	hash := crc32.NewIEEE()
	if _, err := io.Copy(hash, fd); err != nil {
		return "", err
	}
	crc32Val := hash.Sum32()
	return strconv.FormatUint(uint64(crc32Val), 10), nil
}
