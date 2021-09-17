package handler

import (
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/fanghongbo/log-agent/common/g"
	"github.com/fanghongbo/log-agent/utils"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type LogService struct {
	Src         string         `json:"src"`
	Dest        string         `json:"dest"`
	Pattern     string         `json:"pattern"`
	Exclude     string         `json:"exclude"`
	ForbidWrite bool           `json:"forbidWrite"`
	AfterDelete bool           `json:"afterDelete"`
	BeforeTime  int64          `json:"beforeTime"`
	Sign        string         `json:"sign"`
	Spec        string         `json:"spec"`
	Lock        sync.RWMutex   `json:"lock"`
	Running     bool           `json:"running"`
	MathExpr    *regexp.Regexp `json:"-"`
	ExcludeExpr *regexp.Regexp `json:"-"`
	OssClient   *oss.Client    `json:"-"`
}

type FileInfo struct {
	Name string      `json:"name"`
	Info os.FileInfo `json:"info"`
}

func NewFileInfo(name string, info os.FileInfo) *FileInfo {
	return &FileInfo{Name: name, Info: info}
}

func NewLogService(src string, dest string, pattern string, exclude string, forbidWrite bool, afterDelete bool, beforeTime int64, sign string, spec string) *LogService {
	var (
		service *LogService
	)

	service = &LogService{
		Src:         src,
		Dest:        dest,
		Pattern:     pattern,
		Exclude:     exclude,
		ForbidWrite: forbidWrite,
		AfterDelete: afterDelete,
		BeforeTime:  beforeTime,
		Sign:        sign, Spec: spec,
		Lock:        sync.RWMutex{},
		MathExpr:    regexp.MustCompile(pattern),
		ExcludeExpr: regexp.MustCompile(exclude),
	}

	service.OssClient, _ = oss.New(g.Cfg.Config().Endpoint, g.Cfg.Config().AccessKey, g.Cfg.Config().SecretKey)

	return service
}

func (u *LogService) setRunning() {
	u.Lock.Lock()
	defer u.Lock.Unlock()
	u.Running = true
}

func (u *LogService) setCancel() {
	u.Lock.Lock()
	defer u.Lock.Unlock()
	u.Running = false
}

func (u *LogService) isRunning() bool {
	u.Lock.RLock()
	defer u.Lock.RUnlock()
	return u.Running
}

func (u *LogService) Run() {
	var (
		flag  bool
		files []*FileInfo
		err   error
	)

	if u.isRunning() {
		g.AppLog.Warnf("scan path: %v job is in progress", u.Src)
		return
	}

	u.setRunning()
	flag = true

	defer func() {
		if flag {
			u.setCancel()
		}
	}()

	g.AppLog.Infof("begin scan path: %v", u.Src)

	// 遍历备份目录
	if files, err = u.ListDir(); err != nil {
		g.AppLog.Errorf(err.Error())
		return
	}

	// 遍历备份文件
	for _, file := range files {
		if !u.isMatch(strings.TrimPrefix(file.Name, u.Src)) {
			continue
		}

		if u.isExclude(strings.TrimPrefix(file.Name, u.Src)) {
			continue
		}

		if u.BeforeTime != 0 {
			if int64(time.Now().Sub(file.Info.ModTime()).Seconds()) < u.BeforeTime {
				continue
			}
		}

		// 检查是否存在空目录, 如果存在则删除
		if file.Info.IsDir() {
			if err = u.removeDir(file); err != nil {
				g.AppLog.Errorf(err.Error())
			}
		} else {
			if err = u.Backup(file.Name); err != nil {
				g.AppLog.Errorf(err.Error())
			} else {
				g.AppLog.Infof("upload file: %v success", file.Name)
			}
		}
	}
}

func (u *LogService) ListDir() ([]*FileInfo, error) {
	var (
		files []*FileInfo
		err   error
	)

	files = make([]*FileInfo, 0)

	if err = filepath.Walk(u.Src, func(filename string, file os.FileInfo, err error) error { //遍历目录
		if err != nil {
			return err
		}

		files = append(files, NewFileInfo(filename, file))

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func (u *LogService) isMatch(filename string) bool {
	if u.Pattern != "" {
		return u.MathExpr.MatchString(filename)
	} else {
		return true
	}
}

func (u *LogService) isExclude(filename string) bool {
	if u.Exclude != "" {
		return u.ExcludeExpr.MatchString(filename)
	} else {
		return false
	}
}

func (u *LogService) isEmpty(file *FileInfo) (bool, error) {
	var (
		files []fs.FileInfo
		err   error
	)

	if !file.Info.IsDir() {
		return false, fmt.Errorf("%s must be a directory", file.Name)
	}

	if files, err = ioutil.ReadDir(file.Name); err != nil {
		return false, err
	}

	if len(files) == 0 {
		return true, nil
	}

	return false, nil
}

func (u *LogService) removeDir(file *FileInfo) error {
	var (
		isEmpty bool
		err     error
	)

	if isEmpty, err = u.isEmpty(file); err != nil {
		return err
	}

	// 如果目录为空, 则删除这个目录
	if isEmpty {
		if err = os.RemoveAll(file.Name); err != nil {
			return err
		} else {
			return fmt.Errorf("remove empty dir: %v", file.Name)
		}
	}

	return nil
}

func (u *LogService) removeFile(filename string) error {
	var err error

	if err = os.RemoveAll(filename); err != nil {
		return err
	}

	return nil
}

func (u *LogService) Backup(filename string) error {
	var (
		zipName string
		err     error
	)

	// 打包压缩文件
	zipName = fmt.Sprintf("%v.zip", filename)

	if err = utils.Zip(filename, zipName, u.Sign); err != nil {
		return fmt.Errorf("failed to compress file: %v", filename)
	} else {
		// 将压缩好的文件，上传到oss
		if err = u.UploadFile(zipName); err != nil {
			return fmt.Errorf("upload file: %v err: %v", zipName, err.Error())
		} else {
			if err = u.removeFile(zipName); err != nil {
				return fmt.Errorf("remove file: %v err: %v", zipName, err.Error())
			}
		}

		// 上传完成之后删除源文件
		if u.AfterDelete {
			if err = u.removeFile(filename); err != nil {
				return fmt.Errorf("remove file: %v err: %v", filename, err.Error())
			}
		}
	}

	return nil
}

func (u *LogService) UploadFile(filename string) error {
	var (
		bucket    *oss.Bucket
		objectKey string
		options   []oss.Option
		err       error
	)

	if bucket, err = u.OssClient.Bucket(g.Cfg.Config().BucketName); err != nil {
		return err
	}

	objectKey = path.Join(u.Dest, strings.TrimPrefix(filename, u.Src))

	if strings.HasPrefix(objectKey, string(filepath.Separator)) {
		objectKey = strings.TrimPrefix(objectKey, string(filepath.Separator))
	}

	options = []oss.Option{
		oss.Routines(5),
		oss.Checkpoint(true, ""),
		oss.ForbidOverWrite(u.ForbidWrite),
		oss.Meta("HOSTNAME", g.Cfg.Config().Hostname),
	}

	if err = bucket.UploadFile(objectKey, filename, g.Cfg.Config().PartSize, options...); err != nil {
		return u.reformatOssErr(err)
	}

	return nil
}

func (u *LogService) reformatOssErr(err error) error {
	switch err.(type) {
	case oss.ServiceError:
		serviceErr := err.(oss.ServiceError)
		return errors.New(serviceErr.Message)
	default:
		return err
	}
}
